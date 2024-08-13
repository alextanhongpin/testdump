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

	if err := dump(t, b, opts...); err != nil {
		t.Fatal(err)
	}
}

func dump(t *testing.T, b []byte, opts ...Option) (err error) {
	opt := newOptions().apply(opts...)

	var path string
	if opt.file != "" {
		path = filepath.Join("testdata", t.Name(), fmt.Sprintf("%s.txt", opt.file))
	} else {
		path = filepath.Join("testdata", fmt.Sprintf("%s.txt", t.Name()))
	}

	f, err := file.New(path, opt.overwrite())
	if err != nil {
		return err
	}
	defer f.Close()

	return snapshot.Snapshot(f, opt.encoder(), opt.comparer(), b)
}

type encoder struct {
	marshalFns []Transformer
}

func (e *encoder) Marshal(v any) (b []byte, err error) {
	b = v.([]byte)
	for _, transform := range e.marshalFns {
		b, err = transform(b)
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
