package textdump

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alextanhongpin/testdump/pkg/file"
)

type Option func(o *options)

type options struct {
	colors       bool
	env          string
	file         string
	transformers []Transformer
}

func newOptions() *options {
	return &options{
		colors: true,
		env:    "TESTDUMP",
	}
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *options) overwrite() bool {
	t, _ := strconv.ParseBool(os.Getenv(o.env))
	return t
}

func (o *options) encoder() *encoder {
	return &encoder{
		marshalFns: o.transformers,
	}
}

func (o *options) comparer() *comparer {
	return &comparer{colors: o.colors}
}

func (o *options) newReadWriteCloser(name string) (io.ReadWriteCloser, error) {
	path := filepath.Join("testdata", fmt.Sprintf("%s.txt", filepath.Join(name, o.file)))
	return file.New(path, o.overwrite())
}

type Transformer func(b []byte) ([]byte, error)

func Transformers(t ...Transformer) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, t...)
	}
}

func Colors(colors bool) Option {
	return func(o *options) {
		o.colors = colors
	}
}

func Env(env string) Option {
	return func(o *options) {
		o.env = env
	}
}

func File(file string) Option {
	return func(o *options) {
		o.file = file
	}
}
