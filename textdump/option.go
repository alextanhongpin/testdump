package textdump

import (
	"os"
	"strconv"
)

const env = "TESTDUMP"

type Option func(o *options)

type options struct {
	colors       bool
	env          string
	file         string
	transformers []func([]byte) ([]byte, error)
}

func newOptions() *options {
	return &options{
		colors: true,
		env:    env,
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
	return &comparer{
		colors: o.colors,
	}
}

func Transformers(fns ...func([]byte) ([]byte, error)) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, fns...)
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
