package mysqldump

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/alextanhongpin/testdump/pkg/diff"

	"github.com/alextanhongpin/testdump/mysqldump/internal"
)

var d *Dumper

func init() {
	d = new(Dumper)
}

func New(opts ...Option) *Dumper {
	return &Dumper{
		opts: opts,
	}
}

func Dump(t *testing.T, received *SQL, opts ...Option) {
	d.Dump(t, received, opts...)
}

type Dumper struct {
	opts []Option
}

func (d *Dumper) Dump(t *testing.T, received *SQL, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	if err := dump(t, received, opts...); err != nil {
		t.Error(err)
	}
}

func dump(t *testing.T, received *SQL, opts ...Option) error {
	opt := newOption(opts...)
	receivedBytes, err := Write(received, opt.transformers...)
	if err != nil {
		return err
	}

	file := filepath.Join("testdata", fmt.Sprintf("%s.sql", filepath.Join(t.Name(), opt.file)))
	overwrite, _ := strconv.ParseBool(os.Getenv(opt.env))
	written, err := internal.WriteFile(file, receivedBytes, overwrite)
	if err != nil {
		return err
	}

	if written {
		return nil
	}

	snapshotBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	snapshot, err := Read(snapshotBytes)
	if err != nil {
		return err
	}

	comparer := diff.Text
	if opt.colors {
		comparer = diff.ANSI
	}

	if err := snapshot.Compare(received, opt.cmpOpt, comparer); err != nil {
		if opt.file != "" {
			return fmt.Errorf("%s: %w", opt.file, err)
		}

		return err
	}

	return nil
}
