package textdump

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/alextanhongpin/dump/pkg/diff"
	"github.com/alextanhongpin/dump/textdump/internal"
)

var d *dumper

func init() {
	d = new(dumper)
}

type dumper struct {
	opt []Option
}

func Dump(t *testing.T, received []byte, opts ...Option) {
	d.Dump(t, received, opts...)
}

func (d *dumper) Dump(t *testing.T, received []byte, opts ...Option) {
	t.Helper()

	if err := dump(t, received, opts...); err != nil {
		t.Fatal(err)
	}
}

func dump(t *testing.T, received []byte, opts ...Option) error {
	opt := newOption(opts...)

	for _, transform := range opt.transformers {
		var err error
		received, err = transform(received)
		if err != nil {
			return err
		}
	}

	file := fmt.Sprintf("testdata/%s.txt", or(opt.file, t.Name()))
	overwrite, _ := strconv.ParseBool(os.Getenv(opt.env))
	written, err := internal.WriteFile(file, received, overwrite)
	if err != nil {
		return err
	}

	if written {
		return nil
	}

	snapshot, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	comparer := diff.Text
	if opt.colors {
		comparer = diff.ANSI
	}

	// Convert to string for better diff.
	return comparer(string(snapshot), string(received))
}

func or[T comparable](a, b T) T {
	var zero T
	if a != zero {
		return a
	}

	return b
}
