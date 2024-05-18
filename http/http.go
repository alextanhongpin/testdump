package http

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alextanhongpin/dump/http/internal"
)

var update bool

func init() {
	// TODO: Change this to avoid conflict in naming.
	flag.BoolVar(&update, "update", false, "force update the httpdump testdata")
}

type Handler struct {
	t                     *testing.T
	h                     http.Handler
	Middlewares           []Middleware
	CompareRequestOption  *CompareOption
	CompareResponseOption *CompareOption
	// Output.
	SnapshotRequest  *http.Request
	SnapshotResponse *http.Response
	ReceivedRequest  *http.Request
	ReceivedResponse *http.Response
	FS               fs.FS
	PrettyJSON       bool
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

	// do something with w first.
	h.h.ServeHTTP(wr, r)

	// Reset the request.
	r.Body = io.NopCloser(bytes.NewReader(rb))

	if err := h.dump(wr.Result(), r); err != nil {
		t.Fatal(err)
	}
}

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

// apply applies the middleware, cloning the request/response in the process.
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

func (h *Handler) write(file string, wc *http.Response, rc *http.Request) (bool, error) {
	src, err := Write(wc, rc, h.PrettyJSON)
	if err != nil {
		return false, err
	}

	return internal.WriteFile(file, src, update)
}

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

func CompareRequest(s, t *http.Request, opt *CompareOption) error {
	lhs, err := NewComparableRequest(s)
	if err != nil {
		return err
	}

	rhs, err := NewComparableRequest(t)
	if err != nil {
		return err
	}

	return lhs.Compare(rhs, opt)
}

func CompareResponse(s, t *http.Response, opt *CompareOption) error {
	lhs, err := NewComparableResponse(s)
	if err != nil {
		return err
	}

	rhs, err := NewComparableResponse(t)
	if err != nil {
		return err
	}

	return lhs.Compare(rhs, opt)
}
