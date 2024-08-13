package jsondump

import (
	"os"
	"strconv"

	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/google/go-cmp/cmp"
)

// Define a constant for ignored values
const (
	ignoreValue = "[IGNORED]"
)

// Define a function type Option that takes a pointer to an Options struct
type Option func(o *Options)

// Define the Options struct with various fields
type Options struct {
	File                    string // A custom File name.
	Env                     string // The environment variable name to overwrite the snapsnot.
	CmpOpts                 []cmp.Option
	Transformers            []func([]byte) ([]byte, error)
	IgnorePathsTransformers []func([]byte) ([]byte, error)
	Colors                  bool
	Registry                *Registry
}

func (o *Options) overwrite() bool {
	t, _ := strconv.ParseBool(os.Getenv(o.Env))
	return t
}

func (o *Options) apply(opts ...Option) *Options {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

func (o *Options) encoder() *jsonEncoder {
	return &jsonEncoder{
		marshalFns:   o.Transformers,
		unmarshalFns: o.IgnorePathsTransformers,
	}
}

func (o *Options) comparer() *comparer {
	return &comparer{
		cmpOpts: o.CmpOpts,
		colors:  o.Colors,
	}
}

func NewOptions() *Options {
	return &Options{
		Colors: true,
		Env:    "TESTDUMP",
	}
}

// File is an Option that sets the File name
func File(name string) Option {
	return func(o *Options) {
		o.File = name
	}
}

// Env is an Option that sets the environment variable name
func Env(name string) Option {
	return func(o *Options) {
		o.Env = name
	}
}

// Colors is an Option that sets the Colors flag
func Colors(Colors bool) Option {
	return func(o *Options) {
		o.Colors = Colors
	}
}

// Transformer is an Option that adds a transformer function
func Transformer(p ...func([]byte) ([]byte, error)) Option {
	return func(o *Options) {
		o.Transformers = append(o.Transformers, p...)
	}
}

// CmpOpts is an Option that adds comparison options
func CmpOpts(opts ...cmp.Option) Option {
	return func(o *Options) {
		o.CmpOpts = append(o.CmpOpts, opts...)
	}
}

// IgnoreFields is an Option that ignores certain fields
func IgnoreFields(fields ...string) Option {
	return func(o *Options) {
		o.CmpOpts = append(o.CmpOpts, internal.IgnoreMapEntries(fields...))
	}
}

// IgnorePaths is an Option that ignores certain paths
func IgnorePaths(paths ...string) Option {
	return func(o *Options) {
		o.IgnorePathsTransformers = append(o.IgnorePathsTransformers, internal.MaskPaths(ignoreValue, paths))
	}
}

// MaskFields is an Option that masks certain fields
func MaskFields(mask string, fields []string) Option {
	return Transformer(internal.MaskFields(mask, fields))
}

// MaskPaths is an Option that masks certain paths
func MaskPaths(mask string, paths []string) Option {
	return Transformer(internal.MaskPaths(mask, paths))
}

func WithRegistry(reg *Registry) Option {
	return func(o *Options) {
		o.Registry = reg
	}
}

// Define a struct for a Masker
type Masker struct {
	mask string
}

// NewMask is a constructor for the Masker struct
func NewMask(mask string) *Masker {
	return &Masker{mask: mask}
}

// MaskFields is a method on Masker that masks certain fields
func (m *Masker) MaskFields(fields ...string) Option {
	return MaskFields(m.mask, fields)
}

// MaskPaths is a method on Masker that masks certain paths
func (m *Masker) MaskPaths(paths ...string) Option {
	return MaskPaths(m.mask, paths)
}
