package pgdump

import (
	"fmt"
	"io"
	"testing"

	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
)

var d *Dumper

func init() {
	d = New()
}

func Dump(t *testing.T, received *SQL, opts ...Option) {
	d.Dump(t, received, opts...)
}

type Dumper struct {
	opts []Option
}

func New(opts ...Option) *Dumper {
	return &Dumper{
		opts: opts,
	}
}

func (d *Dumper) Dump(t *testing.T, received *SQL, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)

	opt := newOptions().apply(opts...)
	rwc, err := opt.newReadWriteCloser(t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer rwc.Close()

	if err := Snapshot(rwc, received, opts...); err != nil {
		t.Error(err)
	}
}

func Snapshot(rw io.ReadWriter, s *SQL, opts ...Option) error {
	opt := newOptions().apply(opts...)
	return snapshot.Snapshot(rw, opt.encoder(), opt.comparer(), s)
}

type encoder struct {
	marshalFns []func(*SQL) error
}

func (e *encoder) Marshal(v any) ([]byte, error) {
	return Write(v.(*SQL), e.marshalFns...)
}

func (e *encoder) Unmarshal(b []byte) (a any, err error) {
	return Read(b)
}

type comparer struct {
	cmpOpt CompareOption
	colors bool
	file   string
}

func (c *comparer) Compare(a, b any) error {
	x := a.(*SQL)
	y := b.(*SQL)

	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

	err := x.Compare(y, c.cmpOpt, comparer)
	if err != nil {
		if c.file != "" {
			return fmt.Errorf("%s: %w", c.file, err)
		}

		return err
	}

	return nil
}
