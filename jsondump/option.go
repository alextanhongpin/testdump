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
	indent                  bool
	typ                     any
}

// newOption is a constructor for the option struct
func newOption(a any, opts ...Option) *option {
	opt := new(option)
	opt.colors = true
	opt.indent = true
	opt.typ = a

	// Apply each Option function to the new option
	for _, o := range opts {
		o(opt)
	}

	// Add the indent processor as the last processor to pretty print.
	if opt.indent {
		opt.transformers = append(opt.transformers, internal.Indent)
	}

	return opt
}

// MaskPathsFromStructTag is an Option that masks paths from a struct tag
func MaskPathsFromStructTag(key, val, maskValue string) Option {
	return func(o *option) {
		maskPaths := internal.MaskPathsFromStructTag(o.typ, key, val)
		o.transformers = append(o.transformers, internal.MaskPaths(maskValue, maskPaths))
	}
}

// IgnorePathsFromStructTag is an Option that ignores paths from a struct tag
func IgnorePathsFromStructTag(key, val string) Option {
	return func(o *option) {
		maskPaths := internal.IgnorePathsFromStructTag(o.typ, key, val)
		o.ignorePathsTransformers = append(o.ignorePathsTransformers, internal.MaskPaths(ignoreValue, maskPaths))
	}
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

// Indent is an Option that sets the indent flag
func Indent(indent bool) Option {
	return func(o *option) {
		o.indent = indent
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

// Define a struct for a masker
type masker struct {
	mask string
}

// NewMask is a constructor for the masker struct
func NewMask(mask string) *masker {
	return &masker{mask: mask}
}

// MaskFields is a method on masker that masks certain fields
func (m *masker) MaskFields(fields ...string) Option {
	return MaskFields(m.mask, fields)
}

// MaskPaths is a method on masker that masks certain paths
func (m *masker) MaskPaths(paths ...string) Option {
	return MaskPaths(m.mask, paths)
}
