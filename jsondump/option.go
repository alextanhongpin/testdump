package jsondump

import (
	"github.com/alextanhongpin/dump/jsondump/internal"
	"github.com/google/go-cmp/cmp"
)

const (
	ignoreValue = "[IGNORED]"
)

type option struct {
	file                    string // A custom file name.
	env                     string // The environment variable name to overwrite the snapsnot.
	cmpOpts                 []cmp.Option
	transformers            []func([]byte) ([]byte, error)
	ignorePathsTransformers []func([]byte) ([]byte, error)
	colors                  bool
	indent                  bool
}

func newOption(a any, opts ...Option) *option {
	opt := new(option)
	opt.colors = true
	opt.indent = true

	for _, o := range opts {
		switch v := o.(type) {
		case cmpOpts:
			opt.cmpOpts = append(opt.cmpOpts, v...)
		case modifier:
			v(opt)
		case structModifier:
			v(a)(opt)
		}
	}

	// Add the indent processor as the last processor to pretty print.
	if opt.indent {
		opt.transformers = append(opt.transformers, internal.IndentTransformer)
	}

	return opt
}

type Option interface {
	isOption()
}

type modifier func(o *option)

func (m modifier) isOption() {}

type structModifier func(a any) modifier

func (m structModifier) isOption() {}

func MaskPathsFromStructTag(key, val, maskValue string) structModifier {
	return func(a any) modifier {
		maskPaths := internal.MaskPathsFromStructTag(a, key, val)
		return func(o *option) {
			o.transformers = append(o.transformers, internal.MaskPaths(maskValue, maskPaths))
		}
	}
}

func IgnorePathsFromStructTag(key, val string) structModifier {
	return func(a any) modifier {
		maskPaths := internal.IgnorePathsFromStructTag(a, key, val)
		return func(o *option) {
			o.ignorePathsTransformers = append(o.ignorePathsTransformers, internal.MaskPaths(ignoreValue, maskPaths))
		}
	}
}

// File is the json file name.
func File(name string) modifier {
	return func(o *option) {
		o.file = name
	}
}

// Env is the environment variable name to overwrite the snapshot.
func Env(name string) modifier {
	return func(o *option) {
		o.env = name
	}
}

func Colors(colors bool) modifier {
	return func(o *option) {
		o.colors = colors
	}
}

func Indent(indent bool) modifier {
	return func(o *option) {
		o.indent = indent
	}
}

func Transformer(p ...func([]byte) ([]byte, error)) modifier {
	return func(o *option) {
		o.transformers = append(o.transformers, p...)
	}
}

type cmpOpts []cmp.Option

func (cmpOpts) isOption() {}

func CmpOpts(opts ...cmp.Option) cmpOpts {
	return cmpOpts(opts)
}

// IgnoreFields ignores the field names from being compared. All fields with
// the same name, regardless of the path, will be ignored.
func IgnoreFields(fields ...string) modifier {
	return func(o *option) {
		o.cmpOpts = append(o.cmpOpts, internal.IgnoreMapEntries(fields...))
	}
}

// IgnorePaths ignores the field paths from being compared. The path starts
// with the root `$`.
func IgnorePaths(paths ...string) modifier {
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
func MaskFields(mask string, fields []string) modifier {
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
func MaskPaths(mask string, paths []string) modifier {
	return Transformer(internal.MaskPaths(mask, paths))
}

type masker struct {
	mask string
}

func NewMask(mask string) *masker {
	return &masker{mask: mask}
}

func (m *masker) MaskFields(fields ...string) modifier {
	return MaskFields(m.mask, fields)
}

func (m *masker) MaskPaths(paths ...string) modifier {

	return MaskPaths(m.mask, paths)
}
