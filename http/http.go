package http

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/alextanhongpin/dump/http/internal"
)

// Env is the environment variable to enable update mode.
var Env = "TESTDATA_UPDATE"

type Handler struct {
	t *testing.T
	h http.Handler

	// Options.
	Middlewares           []Middleware
	CompareRequestOption  *CompareOption
	CompareResponseOption *CompareOption
	PrettyJSON            bool
	FS                    fs.FS

	// Output.
	SnapshotRequest  *http.Request
	SnapshotResponse *http.Response
	ReceivedRequest  *http.Request
	ReceivedResponse *http.Response
}

func NewHandler(t *testing.T, h http.Handler, middlewares ...Middleware) *Handler {
	return &Handler{
		t:                     t,
		h:                     h,
		Middlewares:           middlewares,
		CompareRequestOption:  new(CompareOption),
		CompareResponseOption: new(CompareOption),
		PrettyJSON:            true,
	}
}

func NewHandlerFunc(t *testing.T, h http.HandlerFunc, middlewares ...Middleware) *Handler {
	return &Handler{
		t:                     t,
		h:                     http.HandlerFunc(h),
		Middlewares:           middlewares,
		CompareRequestOption:  new(CompareOption),
		CompareResponseOption: new(CompareOption),
		PrettyJSON:            true,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := h.t
	t.Helper()

	wr, ok := w.(*httptest.ResponseRecorder)
	if !ok {
		t.Fatal("expected *httptest.ResponseRecorder")
	}

	var rb []byte
	if r.Body != nil {
		defer r.Body.Close()

		var err error
		rb, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		r.Body = io.NopCloser(bytes.NewReader(rb))
	}

	h.h.ServeHTTP(wr, r)

	// Reset the request, which has been read.
	r.Body = io.NopCloser(bytes.NewReader(rb))

	if err := h.dump(wr.Result(), r); err != nil {
		t.Fatal(err)
	}
}

// dump executes the snapshot process.
func (h *Handler) dump(w *http.Response, r *http.Request) error {
	wc, rc, err := h.apply(w, r)
	if err != nil {
		return err
	}

	file := fmt.Sprintf("testdata/%s.http", h.t.Name())
	written, err := h.write(file, wc, rc)
	if err != nil {
		return err
	}

	h.ReceivedRequest = rc
	h.ReceivedResponse = wc

	// First write, there's nothing to compare.
	if written {
		h.SnapshotRequest = rc
		h.SnapshotResponse = wc

		return nil
	}

	ww, rr, err := h.read(file)
	if err != nil {
		return err
	}

	h.SnapshotRequest = rr
	h.SnapshotResponse = ww

	if err := CompareRequest(
		h.SnapshotRequest,
		h.ReceivedRequest,
		h.CompareRequestOption,
	); err != nil {
		return fmt.Errorf("Request %w", err)
	}

	if err := CompareResponse(
		h.SnapshotResponse,
		h.ReceivedResponse,
		h.CompareResponseOption,
	); err != nil {
		return fmt.Errorf("Response %w", err)
	}

	return nil
}

// apply applies the middleware that modifies the request and response.
// The request and response is cloned before the modification, so it does not
// affect the original request or response.
func (h *Handler) apply(w *http.Response, r *http.Request) (*http.Response, *http.Request, error) {
	wc, err := internal.CloneResponse(w)
	if err != nil {
		return nil, nil, err
	}

	rc, err := internal.CloneRequest(r)
	if err != nil {
		return nil, nil, err
	}

	for _, m := range h.Middlewares {
		if err := m(wc, rc); err != nil {
			return nil, nil, err
		}
	}

	return wc, rc, nil
}

// write writes the received data to the file, only if it doesn't exist.
func (h *Handler) write(file string, wc *http.Response, rc *http.Request) (bool, error) {
	src, err := Write(wc, rc, h.PrettyJSON)
	if err != nil {
		return false, err
	}

	update, _ := strconv.ParseBool(os.Getenv(Env))
	return internal.WriteFile(file, src, update)
}

// read reads the snapshot data from the file.
func (h *Handler) read(file string) (*http.Response, *http.Request, error) {
	var tgt []byte
	var err error
	if h.FS != nil {
		tgt, err = fs.ReadFile(h.FS, file)
	} else {
		tgt, err = os.ReadFile(file)
	}
	if err != nil {
		return nil, nil, err
	}

	return Read(tgt)
}
