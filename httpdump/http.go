package httpdump

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"testing"

	"github.com/alextanhongpin/testdump/httpdump/internal"
	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/alextanhongpin/testdump/pkg/file"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
)

// d is a global variable that holds a pointer to a Dumper instance.
var d *Dumper

// init is a special function in Go that is called when the package is initialized.
func init() {
	// We initialize d with a new Dumper instance.
	d = new(Dumper)
}

// Handler is a function that takes a testing object, an HTTP handler, and a variadic list of options.
// It returns an HTTP handler that is wrapped with the Dumper's Handler method.
func Handler(t *testing.T, h http.Handler, opts ...Option) http.Handler {
	return d.Handler(t, h, opts...)
}

// HandlerFunc is similar to Handler, but it takes an HTTP handler function instead of an HTTP handler.
func HandlerFunc(t *testing.T, h http.HandlerFunc, opts ...Option) http.Handler {
	return d.HandlerFunc(t, h, opts...)
}

// Dump is a function that takes a testing object, an HTTP response writer, an HTTP request, and a variadic list of options.
// It calls the Dumper's Dump method with these arguments.
func Dump(t *testing.T, w *http.Response, r *http.Request, opts ...Option) {
	d.Dump(t, w, r, opts...)
}

// Dumper is a struct that holds a slice of options.
type Dumper struct {
	opts []Option
}

// New is a function that takes a variadic list of options and returns a new Dumper instance with these options.
func New(opts ...Option) *Dumper {
	return &Dumper{
		opts: opts,
	}
}

// Handler is a method on the Dumper struct that takes a testing object, an
// HTTP handler, and a variadic list of options. It returns a new handler
// instance with these values.
func (d *Dumper) Handler(t *testing.T, h http.Handler, opts ...Option) http.Handler {
	return &handler{
		t:    t,
		h:    http.Handler(h),
		opts: append(d.opts, opts...),
	}
}

// HandlerFunc is similar to Handler, but it takes an HTTP handler function
// instead of an HTTP handler.
func (d *Dumper) HandlerFunc(t *testing.T, h http.HandlerFunc, opts ...Option) http.Handler {
	return &handler{
		t:    t,
		h:    h,
		opts: append(d.opts, opts...),
	}
}

// Dump is a method on the Dumper struct that takes a testing object, an HTTP response writer, an HTTP request, and a variadic list of options.
// It appends the options to the Dumper's options and then calls the dump function with the testing object, a new HTTP instance, and the options.
func (d *Dumper) Dump(t *testing.T, w *http.Response, r *http.Request, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	if err := dump(t, &HTTP{Request: r, Response: w}, opts...); err != nil {
		t.Error(err)
	}
}

// handler is a struct that holds a testing object, an HTTP handler, and a slice of options.
type handler struct {
	t    *testing.T
	h    http.Handler
	opts []Option
}

// ServeHTTP is a method on the handler struct that takes an HTTP response writer and an HTTP request.
// It clones the request, calls the handler's ServeHTTP method with the response writer and the cloned request,
// and then calls the dump function with the testing object, a new HTTP instance, and the options.
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

// dump is a function that takes a testing object, an HTTP instance, and a variadic list of options.
// It clones the HTTP instance, applies the transformers to the cloned instance, writes the cloned instance to a file,
// reads the snapshot data from the file, and then compares the snapshot data with the cloned instance.
func dump(t *testing.T, h *HTTP, opts ...Option) error {
	opt := newOptions().apply(opts...)

	if opt.body {
		// Read the body only, and write to a separate file.
		typ, _, err := mime.ParseMediaType(h.Response.Header.Get("Content-Type"))
		if err != nil {
			return err
		}
		exts, err := mime.ExtensionsByType(typ)
		if err != nil {
			return err
		}
		sort.Strings(exts)
		ext := exts[len(exts)-1] // Take the longest.
		bodyFile := fmt.Sprintf("testdata/%s%s", t.Name(), ext)
		if opt.file != "" {
			bodyFile = fmt.Sprintf("testdata/%s/%s%s", t.Name(), opt.file, ext)
		}
		b, err := io.ReadAll(h.Response.Body)
		if err != nil {
			return err
		}

		h.Response.Body = io.NopCloser(bytes.NewReader(b))
		_, err = internal.WriteFile(bodyFile, b, true)
		if err != nil {
			return err
		}
	}

	var path string
	if opt.file != "" {
		path = filepath.Join("testdata", t.Name(), fmt.Sprintf("%s.http", opt.file))
	} else {
		path = filepath.Join("testdata", fmt.Sprintf("%s.http", t.Name()))
	}

	f, err := file.New(path, opt.overwrite())
	if err != nil {
		return err
	}
	defer f.Close()

	return snapshot.Snapshot(f, opt.encoder(), opt.comparer(), h)
}

type encoder struct {
	marshalFns []Transformer
	indentJSON bool
}

func (e *encoder) Marshal(v any) ([]byte, error) {
	h := v.(*HTTP)
	hc, err := h.Clone()
	if err != nil {
		return nil, err
	}

	for _, fn := range e.marshalFns {
		if err := fn(hc.Response, hc.Request); err != nil {
			return nil, err
		}
	}

	return Write(hc, e.indentJSON)
}

func (e *encoder) Unmarshal(b []byte) (any, error) {
	return Read(b)
}

type comparer struct {
	colors bool
	cmpOpt CompareOption
}

func (c *comparer) Compare(a, b any) error {
	x := a.(*HTTP)
	y := b.(*HTTP)

	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

	return x.Compare(y, c.cmpOpt, comparer)
}
