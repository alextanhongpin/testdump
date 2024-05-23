package mysqldump

import (
	"github.com/alextanhongpin/testdump/mysqldump/internal"
)

type option struct {
	file         string
	colors       bool
	env          string
	cmpOpt       CompareOption
	transformers []Transformer
}

func newOption(opts ...Option) *option {
	o := new(option)
	o.colors = true

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type Option func(o *option)

func File(file string) Option {
	return func(o *option) {
		o.file = file
	}
}

func Env(env string) Option {
	return func(o *option) {
		o.env = env
	}
}

func Colors(colors bool) Option {
	return func(o *option) {
		o.colors = colors
	}
}

func IgnoreArgs(args ...string) Option {
	return func(o *option) {
		o.cmpOpt.CmpOpts = append(o.cmpOpt.CmpOpts, internal.IgnoreMapEntries(args...))
	}
}

func Transformers(ts ...Transformer) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, ts...)
	}
}
