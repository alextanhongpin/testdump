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

// Define a function type Option that takes a pointer to an options struct
type Option func(o *options)

// Define the options struct with various fields
type options struct {
	cmpOpts                 []cmp.Option
	colors                  bool
	env                     string // The environment variable name to overwrite the snapsnot.
	ignorePathsTransformers []func([]byte) ([]byte, error)
	rawOutput               bool
	registry                *Registry
	transformers            []func([]byte) ([]byte, error)
	file                    string // A custom file name.
}

func (o *options) overwrite() bool {
	t, _ := strconv.ParseBool(os.Getenv(o.env))
	return t
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}

	return o
}

func (o *options) encoder() *encoder {
	return &encoder{
		marshalFns:   o.transformers,
		unmarshalFns: o.ignorePathsTransformers,
	}
}

func (o *options) comparer() *comparer {
	return &comparer{
		cmpOpts: o.cmpOpts,
		colors:  o.colors,
	}
}

func newOptions() *options {
	return &options{
		colors:    true,
		env:       "TESTDUMP",
		rawOutput: true,
	}
}

// File is an Option that sets the file name
func File(name string) Option {
	return func(o *options) {
		o.file = name
	}
}

// Env is an Option that sets the environment variable name
func Env(name string) Option {
	return func(o *options) {
		o.env = name
	}
}

// Colors is an Option that sets the colors flag
func Colors(colors bool) Option {
	return func(o *options) {
		o.colors = colors
	}
}

// Transformers is an Option that adds a transformer function
func Transformers(p ...func([]byte) ([]byte, error)) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, p...)
	}
}

// CmpOpts is an Option that adds comparison options
func CmpOpts(opts ...cmp.Option) Option {
	return func(o *options) {
		o.cmpOpts = append(o.cmpOpts, opts...)
	}
}

// IgnoreFields is an Option that ignores certain fields
func IgnoreFields(fields ...string) Option {
	return func(o *options) {
		o.cmpOpts = append(o.cmpOpts, internal.IgnoreMapEntries(fields...))
	}
}

// IgnorePaths is an Option that ignores certain paths
func IgnorePaths(paths ...string) Option {
	return func(o *options) {
		o.ignorePathsTransformers = append(o.ignorePathsTransformers, internal.MaskPaths(ignoreValue, paths))
	}
}

// MaskFields is an Option that masks certain fields
func MaskFields(mask string, fields []string) Option {
	return Transformers(internal.MaskFields(mask, fields))
}

// MaskPaths is an Option that masks certain paths
func MaskPaths(mask string, paths []string) Option {
	return Transformers(internal.MaskPaths(mask, paths))
}

func WithRegistry(reg *Registry) Option {
	return func(o *options) {
		o.registry = reg
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
