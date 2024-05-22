package httpdump

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/alextanhongpin/dump/httpdump/internal"
)

func Handler(t *testing.T, h http.Handler, opts ...Option) http.Handler {
	return New(opts...).Handler(t, h)
}

func HandlerFunc(t *testing.T, h http.HandlerFunc, opts ...Option) http.Handler {
	return New(opts...).HandlerFunc(t, h)
}

func Dump(t *testing.T, w *http.Response, r *http.Request, opts ...Option) {
	New(opts...).Dump(t, w, r)
}

type dumper struct {
	opts []Option
}

// New returns a dumper with initial options, to be reused for all instance of
// the handler.
func New(opts ...Option) *dumper {
	return &dumper{
		opts: opts,
	}
}

func (d *dumper) Handler(t *testing.T, h http.Handler, opts ...Option) http.Handler {
	opts = append(d.opts, opts...)

	return &handler{
		t:   t,
		h:   http.Handler(h),
		opt: newOption(opts...),
	}
}

func (d *dumper) HandlerFunc(t *testing.T, h http.HandlerFunc, opts ...Option) http.Handler {
	opts = append(d.opts, opts...)

	return &handler{
		t:   t,
		h:   h,
		opt: newOption(opts...),
	}
}

func (d *dumper) Dump(t *testing.T, w *http.Response, r *http.Request, opts ...Option) {
	opts = append(d.opts, opts...)
	h := &handler{
		t:   t,
		opt: newOption(opts...),
	}

	t.Helper()
	if err := h.dump(w, r); err != nil {
		t.Error(err)
	}
}

type handler struct {
	t   *testing.T
	h   http.Handler
	opt *option
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := h.t
	t.Helper()

	wr, ok := w.(*httptest.ResponseRecorder)
	if !ok {
		t.Fatal("expected *httptest.ResponseRecorder")
	}

	// Use the cloned request to avoid modifying the original request.
	rc, err := internal.CloneRequest(r)
	if err != nil {
		t.Fatal(err)
	}

	h.h.ServeHTTP(wr, rc)

	// Dump the request and response to the file.
	if err := h.dump(wr.Result(), r); err != nil {
		t.Fatal(err)
	}
}

// dump executes the snapshot process.
func (h *handler) dump(w *http.Response, r *http.Request) error {
	received := &HTTP{
		Request:  r,
		Response: w,
	}

	var err error
	received, err = h.apply(received)
	if err != nil {
		return err
	}

	file := fmt.Sprintf("testdata/%s.http", h.t.Name())

	// Write the received data to the file.
	written, err := h.write(file, received)
	if err != nil {
		return err
	}

	// First write, there's nothing to compare.
	if written {
		return nil
	}

	// Read the snapshot data from the file.
	snapshot, err := h.read(file)
	if err != nil {
		return err
	}

	return snapshot.Compare(received, h.opt.cmpOpt)
}

// apply applies the middleware that modifies the request and response.
// The request and response is cloned before the modification, so it does not
// affect the original request or response.
func (h *handler) apply(h2p *HTTP) (*HTTP, error) {
	r, err := internal.CloneRequest(h2p.Request)
	if err != nil {
		return nil, err
	}

	w, err := internal.CloneResponse(h2p.Response)
	if err != nil {
		return nil, err
	}

	for _, m := range h.opt.middlewares {
		if err := m(w, r); err != nil {
			return nil, err
		}
	}

	return &HTTP{
		Request:  r,
		Response: w,
	}, nil
}

// write writes the received data to the file, only if it doesn't exist.
func (h *handler) write(file string, h2p *HTTP) (bool, error) {
	src, err := Write(h2p, h.opt.indentJSON)
	if err != nil {
		return false, err
	}

	update, _ := strconv.ParseBool(os.Getenv(h.opt.env))
	return internal.WriteFile(file, src, update)
}

// read reads the snapshot data from the file.
func (h *handler) read(file string) (*HTTP, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return Read(b)
}
