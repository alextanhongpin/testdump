package jsondump

import (
	gocmp "cmp"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/alextanhongpin/testdump/pkg/file"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
)

var d *Dumper

func init() {
	d = new(Dumper)
}

// New creates a new Dumper with the given options.
// The Dumper can be used to dump values to a file.
func New(opts ...Option) *Dumper {
	return &Dumper{opts: opts}
}

func Dump(t *testing.T, v any, opts ...Option) {
	d.Dump(t, v, opts...)
}

type Dumper struct {
	opts []Option
}

func (d *Dumper) Dump(t *testing.T, v any, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	if err := Snapshot(t, v, opts...); err != nil {
		t.Error(err)
	}
}

func Snapshot(t *testing.T, v any, opts ...Option) error {
	t.Helper()

	opt := NewOptions().apply(opts...)
	if opt.Registry != nil {
		opt.apply(opt.Registry.Get(v)...)
	}

	name := gocmp.Or(opt.File, internal.TypeName(v))
	path := filepath.Join("testdata", t.Name(), fmt.Sprintf("%s.json", name))

	f, err := file.New(path, opt.overwrite())
	if err != nil {
		return err
	}
	defer f.Close()

	return snapshot.Snapshot(f, f, opt.encoder(), v, opt.comparer().Compare)
}

type jsonEncoder struct {
	marshalFns   []func([]byte) ([]byte, error)
	unmarshalFns []func([]byte) ([]byte, error)
}

func (e *jsonEncoder) Marshal(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return nil, err
	}

	for _, fn := range e.marshalFns {
		b, err = fn(b)
		if err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (e *jsonEncoder) Unmarshal(b []byte) (a any, err error) {
	for _, fn := range e.unmarshalFns {
		b, err = fn(b)
		if err != nil {
			return nil, err
		}
	}

	// Convert back to map[string]any for nicer diff.
	err = json.Unmarshal(b, &a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

type comparer struct {
	colors  bool
	cmpOpts []cmp.Option
}

func (c *comparer) Compare(a, b any) error {

	// Since google's cmp does not have an option to ignore paths, we just mask
	// the values before comparing.
	// The masked values will not be written to the file.

	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

	return comparer(a, b, c.cmpOpts...)
}
