package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/alextanhongpin/dump/http/internal"
)

// Env is the environment variable to enable update mode.
var Env = "TESTDATA_UPDATE"

type handler struct {
	t   *testing.T
	h   http.Handler
	opt *option
}

func Handler(t *testing.T, h http.Handler, opts ...Option) *handler {
	return &handler{
		t:   t,
		h:   h,
		opt: newOption(opts...),
	}
}

func HandlerFunc(t *testing.T, h http.HandlerFunc, opts ...Option) *handler {
	return &handler{
		t:   t,
		h:   http.HandlerFunc(h),
		opt: newOption(opts...),
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := h.t
	t.Helper()

	wr, ok := w.(*httptest.ResponseRecorder)
	if !ok {
		t.Fatal("expected *httptest.ResponseRecorder")
	}

	rc, err := internal.CloneRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	h.h.ServeHTTP(wr, rc)

	if err := h.dump(wr.Result(), r); err != nil {
		t.Fatal(err)
	}
}

// dump executes the snapshot process.
func (h *handler) dump(w *http.Response, r *http.Request) error {
	wc, rc, err := h.apply(w, r)
	if err != nil {
		return err
	}

	file := fmt.Sprintf("testdata/%s.http", h.t.Name())
	// Write the received data to the file.
	written, err := h.write(file, wc, rc)
	if err != nil {
		return err
	}

	// First write, there's nothing to compare.
	if written {
		return nil
	}

	// Read the snapshot data from the file.
	ww, rr, err := h.read(file)
	if err != nil {
		return err
	}

	if err := CompareRequest(rr, rc, h.opt.req); err != nil {
		return fmt.Errorf("Request %w", err)
	}

	if err := CompareResponse(ww, wc, h.opt.res); err != nil {
		return fmt.Errorf("Response %w", err)
	}

	return nil
}

// apply applies the middleware that modifies the request and response.
// The request and response is cloned before the modification, so it does not
// affect the original request or response.
func (h *handler) apply(w *http.Response, r *http.Request) (*http.Response, *http.Request, error) {
	wc, err := internal.CloneResponse(w)
	if err != nil {
		return nil, nil, err
	}

	rc, err := internal.CloneRequest(r)
	if err != nil {
		return nil, nil, err
	}

	for _, m := range h.opt.mws {
		if err := m(wc, rc); err != nil {
			return nil, nil, err
		}
	}

	return wc, rc, nil
}

// write writes the received data to the file, only if it doesn't exist.
func (h *handler) write(file string, wc *http.Response, rc *http.Request) (bool, error) {
	src, err := Write(wc, rc, h.opt.indentJSON)
	if err != nil {
		return false, err
	}

	update, _ := strconv.ParseBool(os.Getenv(Env))
	return internal.WriteFile(file, src, update)
}

// read reads the snapshot data from the file.
func (h *handler) read(file string) (*http.Response, *http.Request, error) {
	var tgt []byte
	var err error
	tgt, err = os.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}

	return Read(tgt)
}
