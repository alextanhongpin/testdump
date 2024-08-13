package pgdump

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alextanhongpin/testdump/pgdump/internal"
	"github.com/alextanhongpin/testdump/pkg/file"
)

type options struct {
	cmpOpt       CompareOption
	colors       bool
	env          string
	file         string
	transformers []func(*SQL) error
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

func (o *options) newReadWriteCloser(name string) (io.ReadWriteCloser, error) {
	path := filepath.Join("testdata", fmt.Sprintf("%s.sql", filepath.Join(name, o.file)))
	return file.New(path, o.overwrite())
}

func (o *options) comparer() *comparer {
	return &comparer{
		cmpOpt: o.cmpOpt,
		colors: o.colors,
	}
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
		o.cmpOpt.CmpOpts = append(o.cmpOpt.CmpOpts, internal.IgnoreMapEntries(args...))
	}
}

func Transformers(ts ...func(*SQL) error) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, ts...)
	}
}
