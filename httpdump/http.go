package httpdump

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/alextanhongpin/testdump/httpdump/internal"
	"github.com/alextanhongpin/testdump/pkg/diff"
)

var d *dumper

func init() {
	d = new(dumper)
}

func Handler(t *testing.T, h http.Handler, opts ...Option) http.Handler {
	return d.Handler(t, h, opts...)
}

func HandlerFunc(t *testing.T, h http.HandlerFunc, opts ...Option) http.Handler {
	return d.HandlerFunc(t, h, opts...)
}

func Dump(t *testing.T, w *http.Response, r *http.Request, opts ...Option) {
	d.Dump(t, w, r, opts...)
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
	return &handler{
		t:    t,
		h:    http.Handler(h),
		opts: append(d.opts, opts...),
	}
}

func (d *dumper) HandlerFunc(t *testing.T, h http.HandlerFunc, opts ...Option) http.Handler {
	return &handler{
		t:    t,
		h:    h,
		opts: append(d.opts, opts...),
	}
}

func (d *dumper) Dump(t *testing.T, w *http.Response, r *http.Request, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	if err := dump(t, &HTTP{Request: r, Response: w}, opts...); err != nil {
		t.Error(err)
	}
}

type handler struct {
	t    *testing.T
	h    http.Handler
	opts []Option
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
	if err := dump(t, &HTTP{
		Response: wr.Result(),
		Request:  r,
	}, h.opts...); err != nil {
		t.Fatal(err)
	}
}

// dump executes the snapshot process.
func dump(t *testing.T, h2p *HTTP, opts ...Option) error {
	opt := newOption(opts...)

	received, err := h2p.Clone()
	if err != nil {
		return err
	}

	for _, transform := range opt.transformers {
		if err := transform(received.Response, received.Request); err != nil {
			return err
		}
	}

	file := fmt.Sprintf("testdata/%s.http", t.Name())
	src, err := Write(received, opt.indentJSON)
	if err != nil {
		return err
	}

	update, _ := strconv.ParseBool(os.Getenv(opt.env))

	// Write the received data to the file.
	written, err := internal.WriteFile(file, src, update)
	if err != nil {
		return err
	}

	// First write, there's nothing to compare.
	if written {
		return nil
	}

	// Read the snapshot data from the file.
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	snapshot, err := Read(b)
	if err != nil {
		return err
	}

	comparer := diff.Text
	if opt.colors {
		comparer = diff.ANSI
	}

	return snapshot.Compare(received, opt.cmpOpt, comparer)
}
