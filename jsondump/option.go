package jsondump

import (
	"github.com/alextanhongpin/testdump/jsondump/internal"
	"github.com/google/go-cmp/cmp"
)

const (
	ignoreValue = "[IGNORED]"
)

type Option func(o *option)

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

func newOption(a any, opts ...Option) *option {
	opt := new(option)
	opt.colors = true
	opt.indent = true
	opt.typ = a

	for _, o := range opts {
		o(opt)
	}

	// Add the indent processor as the last processor to pretty print.
	if opt.indent {
		opt.transformers = append(opt.transformers, internal.Indent)
	}

	return opt
}

func MaskPathsFromStructTag(key, val, maskValue string) Option {
	return func(o *option) {
		maskPaths := internal.MaskPathsFromStructTag(o.typ, key, val)
		o.transformers = append(o.transformers, internal.MaskPaths(maskValue, maskPaths))
	}
}

func IgnorePathsFromStructTag(key, val string) Option {
	return func(o *option) {
		maskPaths := internal.IgnorePathsFromStructTag(o.typ, key, val)
		o.ignorePathsTransformers = append(o.ignorePathsTransformers, internal.MaskPaths(ignoreValue, maskPaths))
	}
}

// File is the json file name.
func File(name string) Option {
	return func(o *option) {
		o.file = name
	}
}

// Env is the environment variable name to overwrite the snapshot.
func Env(name string) Option {
	return func(o *option) {
		o.env = name
	}
}

func Colors(colors bool) Option {
	return func(o *option) {
		o.colors = colors
	}
}

func Indent(indent bool) Option {
	return func(o *option) {
		o.indent = indent
	}
}

func Transformer(p ...func([]byte) ([]byte, error)) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, p...)
	}
}

func CmpOpts(opts ...cmp.Option) Option {
	return func(o *option) {
		o.cmpOpts = append(o.cmpOpts, opts...)
	}
}

// IgnoreFields ignores the field names from being compared. All fields with
// the same name, regardless of the path, will be ignored.
func IgnoreFields(fields ...string) Option {
	return func(o *option) {
		o.cmpOpts = append(o.cmpOpts, internal.IgnoreMapEntries(fields...))
	}
}

// IgnorePaths ignores the field paths from being compared. The path starts
// with the root `$`.
func IgnorePaths(paths ...string) Option {
	return func(o *option) {
		o.ignorePathsTransformers = append(o.ignorePathsTransformers, internal.MaskPaths(ignoreValue, paths))
	}
}

// MaskFields masks the field names. It does not take into
// consideration the path.
// For example, for a json with the shape
//
//	{
//		"name": "John",
//		"account": {
//	    "name": "email"
//	  }
//	}
//
// Masking the field `name` would result in both `name` fields being masked.
func MaskFields(mask string, fields []string) Option {
	return Transformer(internal.MaskFields(mask, fields))
}

// MaskPaths masks the field paths. The path starts with the root `$`.
// For example, for a json with the shape
//
//	{
//		"name": "John",
//		"account": {
//	    "email": "john.doe@mail.com"
//	  }
//	}
//
// The path for the name field would be `$.name`.
// And the path for the email field would be `$.account.email`.
func MaskPaths(mask string, paths []string) Option {
	return Transformer(internal.MaskPaths(mask, paths))
}

type masker struct {
	mask string
}

func NewMask(mask string) *masker {
	return &masker{mask: mask}
}

func (m *masker) MaskFields(fields ...string) Option {
	return MaskFields(m.mask, fields)
}

func (m *masker) MaskPaths(paths ...string) Option {

	return MaskPaths(m.mask, paths)
}
