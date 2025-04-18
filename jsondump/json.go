package jsondump

import (
	"bytes"
	gocmp "cmp"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/alextanhongpin/testdump/pkg/file"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
)

var d *Dumper

func init() {
	d = New()
}

func Dump(t *testing.T, v any, opts ...Option) {
	d.Dump(t, v, opts...)
}

func Register(v any, opts ...Option) {
	d.Register(v, opts...)
}

// New creates a new Dumper with the given options.
// The Dumper can be used to dump values to a file.
func New(opts ...Option) *Dumper {
	return &Dumper{
		opts:     opts,
		registry: NewRegistry(),
	}
}

type Dumper struct {
	opts     []Option
	registry *Registry
}

func (d *Dumper) Register(v any, opts ...Option) {
	d.registry.Register(v, opts...)
}

func (d *Dumper) Dump(t *testing.T, v any, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	opts = append(d.registry.Get(v), opts...)
	if err := dump(t, v, opts...); err != nil {
		t.Error(err)
	}
}

func dump(t *testing.T, v any, opts ...Option) error {
	opt := newOptions().apply(opts...)

	name := gocmp.Or(opt.file, internal.TypeName(v))
	path := filepath.Join("testdata", t.Name(), fmt.Sprintf("%s.json", name))
	f, err := file.New(path, opt.overwrite())
	if err != nil {
		return err
	}

	defer f.Close()

	if opt.rawOutput {
		path := filepath.Join("testdata", t.Name(), fmt.Sprintf("%s.out", name))
		o, err := file.New(path, true)
		if err != nil {
			return err
		}
		defer o.Close()

		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}

		_, err = o.Write(b)
		if err != nil {
			return err
		}
	}

	return snapshot.Snapshot(f, opt.encoder(), opt.comparer(), v)
}

type encoder struct {
	marshalFns   []func([]byte) ([]byte, error)
	unmarshalFns []func([]byte) ([]byte, error)
}

func (e *encoder) Marshal(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	for _, fn := range e.marshalFns {
		b, err = fn(b)
		if err != nil {
			return nil, err
		}
	}

	return indent(b)
}

func (e *encoder) Unmarshal(b []byte) (a any, err error) {
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
	colors      bool
	ignorePaths []string
	opts        []cmp.Option
}

func (c *comparer) Compare(a, b any) error {
	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}
	// Before we decide to ignore the paths for comparison, we ensure that
	// - the path is not empty on either side
	// - the value is not empty on either side
	// - the value type is the same on both sides
	aVals := internal.GetMapValues(a, c.ignorePaths...)
	bVals := internal.GetMapValues(b, c.ignorePaths...)
	var remove []string
	if len(aVals) == len(bVals) && len(aVals) == len(c.ignorePaths) {
		for key, aVal := range aVals {
			bVal := bVals[key]
			if aVal == nil || bVal == nil {
				continue
			}
			if reflect.TypeOf(aVal) != reflect.TypeOf(bVal) {
				continue
			}
			remove = append(remove, key)
		}
	}

	ac, err := internal.DeleteMapValues(a, remove...)
	if err != nil {
		return err
	}

	bc, err := internal.DeleteMapValues(b, remove...)
	if err != nil {
		return err
	}

	return comparer(ac, bc, c.opts...)
}

func indent(b []byte) ([]byte, error) {
	if !json.Valid(b) {
		return b, nil
	}

	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
