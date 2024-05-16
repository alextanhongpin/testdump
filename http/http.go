package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type RequestInterceptor Interceptor[*http.Request]
type ResponseInterceptor Interceptor[*http.Response]

type Handler struct {
	t                *testing.T
	h                http.Handler
	Request          RequestInterceptor
	Response         ResponseInterceptor
	RequestComparer  Comparer
	ResponseComparer Comparer
}

type Comparer struct {
	Header  []cmp.Option
	Body    []cmp.Option
	Trailer []cmp.Option
}

func NewHandler(t *testing.T, h http.Handler) *Handler {
	return &Handler{t: t, h: h}
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

	_ = rb // pipe requests
	//_ = wb // pipe response

	if err := h.dump(wr.Result(), r); err != nil {
		t.Fatal(err)
	}
}

func (h *Handler) dump(w *http.Response, r *http.Request) error {
	t := h.t
	t.Helper()
	// comparators
	// do something with rb and wb
	// write response
	src, err := Write(w, r)
	if err != nil {
		return err
	}

	file := fmt.Sprintf("testdata/%s.http", t.Name())
	if err := WriteFile(file, src, false); err != nil {
		return err
	}

	tgt, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	ww, rr, err := Read(tgt)
	if err != nil {
		return err
	}

	{
		// compare request
		lhs, err := DumpRequest(r)
		if err != nil {
			return err
		}

		rhs, err := DumpRequest(rr)
		if err != nil {
			return err
		}

		if err := h.compare("request", lhs, rhs); err != nil {
			return fmt.Errorf("Request %w", err)
		}
	}

	{
		// compare response
		lhs, err := DumpResponse(w)
		if err != nil {
			return err
		}

		rhs, err := DumpResponse(ww)
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

	if err := ANSIDiff(x.Line, y.Line); err != nil {
		return fmt.Errorf("Line: %w", err)
	}

	if err := ANSIDiff(x.Body, y.Body, c.Body...); err != nil {
		return fmt.Errorf("Body: %w", err)
	}

	if err := ANSIDiff(x.Header, y.Header, c.Header...); err != nil {
		return fmt.Errorf("Header: %w", err)
	}

	if err := ANSIDiff(x.Trailer, y.Trailer, c.Trailer...); err != nil {
		return fmt.Errorf("Trailer: %w", err)
	}

	return nil
}
