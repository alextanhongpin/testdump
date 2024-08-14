package yamldump

import (
	gocmp "cmp"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/alextanhongpin/testdump/pkg/file"
	"github.com/alextanhongpin/testdump/pkg/snapshot"
	"github.com/alextanhongpin/testdump/yamldump/internal"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

var d *Dumper

func init() {
	d = New()
}

type Dumper struct {
	opts     []Option
	registry *Registry
}

// New creates a new Dumper with the given options.
// The Dumper can be used to dump values to a file.
func New(opts ...Option) *Dumper {
	return &Dumper{
		opts:     opts,
		registry: NewRegistry(),
	}
}

func Dump(t *testing.T, v any, opts ...Option) {
	d.Dump(t, v, opts...)
}

func Register(v any, opts ...Option) {
	d.Register(v, opts...)
}

func (d *Dumper) Dump(t *testing.T, v any, opts ...Option) {
	t.Helper()

	opts = append(d.opts, opts...)
	opts = append(d.registry.Get(v), opts...)
	if err := dump(t, v, opts...); err != nil {
		t.Error(err)
	}
}

func (d *Dumper) Register(v any, opts ...Option) {
	d.registry.Register(v, opts...)
}

func dump(t *testing.T, v any, opts ...Option) error {
	opt := newOptions().apply(opts...)

	name := gocmp.Or(opt.file, internal.TypeName(v))
	path := filepath.Join("testdata", t.Name(), fmt.Sprintf("%s.yaml", name))
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

		b, err := yaml.Marshal(v)
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

	// Before writing, convert the json back to yaml.
	var a any
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}

	return yaml.Marshal(a)
}

func (e *encoder) Unmarshal(b []byte) (a any, err error) {
	if err := yaml.Unmarshal(b, &a); err != nil {
		return nil, err
	}

	b, err = json.Marshal(a)
	if err != nil {
		return nil, err
	}

	for _, fn := range e.unmarshalFns {
		b, err = fn(b)
		if err != nil {
			return nil, err
		}
	}

	a = nil

	// Convert back to map[string]any for nicer diff.
	err = json.Unmarshal(b, &a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

type comparer struct {
	colors bool
	opts   []cmp.Option
}

func (c *comparer) Compare(a, b any) error {
	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

	return comparer(a, b, c.opts...)
}
