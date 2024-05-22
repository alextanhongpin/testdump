package jsondump

import (
	"github.com/alextanhongpin/dump/jsondump/internal"
	"github.com/google/go-cmp/cmp"
)

const (
	maskValue   = "[REDACTED]"
	ignoreValue = "[IGNORED]"
)

type option struct {
	File                 string // A custom file name.
	Env                  string // The environment variable name to overwrite the snapsnot.
	CmpOpts              []cmp.Option
	Processors           []func([]byte) ([]byte, error)
	IgnorePathsProcessor []func([]byte) ([]byte, error)
}

func newOption(opts ...Option) *option {
	opt := new(option)

	for _, o := range opts {
		switch v := o.(type) {
		case cmpOpts:
			opt.CmpOpts = append(opt.CmpOpts, v...)
		case modifier:
			v(opt)
		}
	}

	// Add the indent processor as the last processor to pretty print.
	opt.Processors = append(opt.Processors, internal.IndentProcessor)

	return opt
}

type Option interface {
	isOption()
}

type modifier func(o *option)

func (m modifier) isOption() {}

// File is the json file name.
func File(name string) modifier {
	return func(o *option) {
		o.File = name
	}
}

// Env is the environment variable name to overwrite the snapshot.
func Env(name string) modifier {
	return func(o *option) {
		o.Env = name
	}
}

func Processor(p ...func([]byte) ([]byte, error)) modifier {
	return func(o *option) {
		o.Processors = append(o.Processors, p...)
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
		o.CmpOpts = append(o.CmpOpts, internal.IgnoreMapEntries(fields...))
	}
}

// IgnorePaths ignores the field paths from being compared. The path starts
// with the root `$`.
func IgnorePaths(paths ...string) modifier {
	return func(o *option) {
		o.IgnorePathsProcessor = append(o.IgnorePathsProcessor, internal.MaskPaths(ignoreValue, paths...))
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
func MaskFields(mask string, fields ...string) modifier {
	return Processor(internal.MaskFields(mask, fields...))
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
func MaskPaths(mask string, paths ...string) modifier {
	return Processor(internal.MaskPaths(mask, paths...))
}
