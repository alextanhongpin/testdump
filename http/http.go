package http

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/alextanhongpin/dump/http/internal"
)

type Handler struct {
	t                *testing.T
	h                http.Handler
	Middlewares      []Middleware
	RequestComparer  *CompareOption
	ResponseComparer *CompareOption
	FS               fs.FS
}

func NewHandler(t *testing.T, h http.Handler, middlewares ...Middleware) *Handler {
	return &Handler{
		t:                t,
		h:                h,
		Middlewares:      middlewares,
		RequestComparer:  new(CompareOption),
		ResponseComparer: new(CompareOption),
	}
}

func NewHandlerFunc(t *testing.T, h http.HandlerFunc, middlewares ...Middleware) *Handler {
	return &Handler{
		t:                t,
		h:                http.HandlerFunc(h),
		Middlewares:      middlewares,
		RequestComparer:  new(CompareOption),
		ResponseComparer: new(CompareOption),
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
	t := h.t
	t.Helper()

	rc, err := internal.CloneRequest(r)
	if err != nil {
		return err
	}

	wc, err := internal.CloneResponse(w)
	if err != nil {
		return err
	}

	for _, m := range h.Middlewares {
		if err := m(wc, rc); err != nil {
			return err
		}
	}

	// comparators
	// do something with rb and wb
	// write response
	src, err := Write(wc, rc)
	if err != nil {
		return err
	}

	file := fmt.Sprintf("testdata/%s.http", t.Name())
	written, err := internal.WriteFile(file, src, false)
	if err != nil {
		return err
	}

	// First write, there's nothing to compare.
	if written {
		return nil
	}

	var tgt []byte
	if h.FS != nil {
		tgt, err = fs.ReadFile(h.FS, file)
	} else {
		tgt, err = os.ReadFile(file)
	}
	if err != nil {
		return err
	}

	ww, rr, err := Read(tgt)
	if err != nil {
		return err
	}

	{
		// compare request
		lhs, err := NewRequestDump(rc)
		if err != nil {
			return err
		}

		rhs, err := NewRequestDump(rr)
		if err != nil {
			return err
		}

		if err := h.compare("request", lhs, rhs); err != nil {
			return fmt.Errorf("Request %w", err)
		}
	}

	{
		// compare response
		lhs, err := NewResponseDump(wc)
		if err != nil {
			return err
		}

		rhs, err := NewResponseDump(ww)
		if err != nil {
			return err
		}

		if err := h.compare("response", lhs, rhs); err != nil {
			return fmt.Errorf("Response %w", err)
		}
	}

	return nil
}

func (h *Handler) compare(requestOrResponse string, snapshot, received *Dump) error {
	x := snapshot
	y := received
	c := h.ResponseComparer
	if requestOrResponse == "request" {
		c = h.RequestComparer
	}
	return x.Compare(y, c)
}
