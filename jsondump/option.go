package jsondump

import (
	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/google/go-cmp/cmp"
)

// Define a constant for ignored values
const (
	ignoreValue = "[IGNORED]"
)

// Define a function type Option that takes a pointer to an option struct
type Option func(o *option)

// Define the option struct with various fields
type option struct {
	file                    string // A custom file name.
	env                     string // The environment variable name to overwrite the snapsnot.
	cmpOpts                 []cmp.Option
	transformers            []func([]byte) ([]byte, error)
	ignorePathsTransformers []func([]byte) ([]byte, error)
	colors                  bool
}

// newOption is a constructor for the option struct
func newOption(opts ...Option) *option {
	opt := new(option)
	opt.colors = true
	opt.env = "TESTDUMP"

	// Apply each Option function to the new option
	for _, o := range opts {
		o(opt)
	}

	return opt
}

// File is an Option that sets the file name
func File(name string) Option {
	return func(o *option) {
		o.file = name
	}
}

// Env is an Option that sets the environment variable name
func Env(name string) Option {
	return func(o *option) {
		o.env = name
	}
}

// Colors is an Option that sets the colors flag
func Colors(colors bool) Option {
	return func(o *option) {
		o.colors = colors
	}
}

// Transformer is an Option that adds a transformer function
func Transformer(p ...func([]byte) ([]byte, error)) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, p...)
	}
}

// CmpOpts is an Option that adds comparison options
func CmpOpts(opts ...cmp.Option) Option {
	return func(o *option) {
		o.cmpOpts = append(o.cmpOpts, opts...)
	}
}

// IgnoreFields is an Option that ignores certain fields
func IgnoreFields(fields ...string) Option {
	return func(o *option) {
		o.cmpOpts = append(o.cmpOpts, internal.IgnoreMapEntries(fields...))
	}
}

// IgnorePaths is an Option that ignores certain paths
func IgnorePaths(paths ...string) Option {
	return func(o *option) {
		o.ignorePathsTransformers = append(o.ignorePathsTransformers, internal.MaskPaths(ignoreValue, paths))
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
