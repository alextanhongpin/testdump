package textdump

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/alextanhongpin/testdump/pkg/file"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
)

var d *Dumper

func init() {
	d = New()
}

type Dumper struct {
	opts []Option
}

func New(opts ...Option) *Dumper {
	return &Dumper{opts: opts}
}

func Dump(t *testing.T, b []byte, opts ...Option) {
	d.Dump(t, b, opts...)
}

func (d *Dumper) Dump(t *testing.T, b []byte, opts ...Option) {
	t.Helper()

	opt := newOptions().apply(append(d.opts, opts...)...)

	path := filepath.Join("testdata", fmt.Sprintf("%s.txt", filepath.Join(t.Name(), opt.file)))
	f, err := file.New(path, opt.overwrite())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := snapshot.Snapshot(f, opt.encoder(), opt.comparer(), b); err != nil {
		t.Fatal(err)
	}
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

type encoder struct {
	marshalFns []func([]byte) ([]byte, error)
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
