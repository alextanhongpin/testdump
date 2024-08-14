package mysqldump

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alextanhongpin/testdump/pkg/file"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
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

func Dump(t *testing.T, s *SQL, opts ...Option) {
	d.Dump(t, s, opts...)
}

type Dumper struct {
	opts []Option
}

func (d *Dumper) Dump(t *testing.T, s *SQL, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	if err := dump(t, s, opts...); err != nil {
		t.Error(err)
	}
}

func dump(t *testing.T, s *SQL, opts ...Option) error {
	opt := newOptions().apply(opts...)

	path := filepath.Join("testdata", fmt.Sprintf("%s.sql", filepath.Join(t.Name(), opt.file)))
	f, err := file.New(path, opt.overwrite())
	if err != nil {
		return err
	}
	defer f.Close()

	return snapshot.Snapshot(f, opt.encoder(), opt.comparer(), s)
}
