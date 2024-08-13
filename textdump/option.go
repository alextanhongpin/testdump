package textdump

import (
	"os"
	"strconv"
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
