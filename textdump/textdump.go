package textdump

import (
	"io"
	"testing"

	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
)

var d *Dumper

func init() {
	d = New()
}

type Dumper struct {
	opt []Option
}

func New(opts ...Option) *Dumper {
	return &Dumper{opt: opts}
}

func Dump(t *testing.T, b []byte, opts ...Option) {
	d.Dump(t, b, opts...)
}

func (d *Dumper) Dump(t *testing.T, b []byte, opts ...Option) {
	t.Helper()

	opts = append(d.opt, opts...)

	opt := newOptions().apply(opts...)
	rwc, err := opt.newReadWriteCloser(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer rwc.Close()

	if err := Snapshot(rwc, b, opts...); err != nil {
		t.Fatal(err)
	}
}

func Snapshot(rw io.ReadWriter, b []byte, opts ...Option) (err error) {
	opt := newOptions().apply(opts...)

	return snapshot.Snapshot(rw, opt.encoder(), opt.comparer(), b)
}

type encoder struct {
	marshalFns []Transformer
}

func (e *encoder) Marshal(v any) (b []byte, err error) {
	b = v.([]byte)
	for _, fn := range e.marshalFns {
		b, err = fn(b)
		if err != nil {
			return
		}
	}

	return
}

func (e *encoder) Unmarshal(b []byte) (a any, err error) {
	return b, nil
}

type comparer struct {
	colors bool
}

func (c *comparer) Compare(a, b any) error {
	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

	// Convert to string for better diff.
	return comparer(string(a.([]byte)), string(b.([]byte)))
}
