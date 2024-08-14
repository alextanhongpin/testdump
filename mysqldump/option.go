package mysqldump

import (
	"os"
	"strconv"

	"github.com/alextanhongpin/testdump/mysqldump/internal"
	"github.com/google/go-cmp/cmp"
)

const env = "TESTDUMP"

type options struct {
	file         string
	colors       bool
	env          string
	cmpOpts      []cmp.Option
	transformers []func(*SQL) error
}

func newOptions() *options {
	return &options{
		colors: true,
		env:    env,
	}
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
		opts:   o.cmpOpts,
		colors: o.colors,
	}
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type Option func(o *options)

func File(file string) Option {
	return func(o *options) {
		o.file = file
	}
}

func Env(env string) Option {
	return func(o *options) {
		o.env = env
	}
}

func Colors(colors bool) Option {
	return func(o *options) {
		o.colors = colors
	}
}

func IgnoreArgs(args ...string) Option {
	return func(o *options) {
		o.cmpOpts = append(o.cmpOpts, internal.IgnoreMapEntries(args...))
	}
}

func Transformers(ts ...func(*SQL) error) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, ts...)
	}
}
