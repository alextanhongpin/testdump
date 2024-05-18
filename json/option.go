package json

import (
	"slices"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const (
	maskValue   = "[REDACTED]"
	ignoreValue = "[IGNORED]"
)

type option struct {
	Name                 string
	CmpOpts              []cmp.Option
	Processors           []Processor
	IgnorePathsProcessor []Processor
	IgnoreFields         []string
	IgnorePaths          []string
	MaskFields           []string
	MaskPaths            []string
}

func newOption(opts ...Option) *option {
	opt := new(option)

	for _, o := range opts {
		switch v := o.(type) {
		case Processor:
			opt.Processors = append(opt.Processors, v)
		case processors:
			opt.Processors = append(opt.Processors, v...)
		case cmpOpts:
			opt.CmpOpts = append(opt.CmpOpts, v...)
		case ignoreFields:
			opt.IgnoreFields = append(opt.IgnoreFields, v...)
		case ignorePaths:
			opt.IgnorePaths = append(opt.IgnorePaths, v...)
		case maskFields:
			opt.MaskFields = append(opt.MaskFields, v...)
		case maskPaths:
			opt.MaskPaths = append(opt.MaskPaths, v...)
		case Name:
			opt.Name = string(v)
		}
	}
	slices.Sort(opt.IgnoreFields)
	slices.Sort(opt.IgnorePaths)
	slices.Sort(opt.MaskFields)
	slices.Sort(opt.MaskPaths)
	opt.IgnoreFields = slices.Compact(opt.IgnoreFields)
	opt.IgnorePaths = slices.Compact(opt.IgnorePaths)
	opt.MaskFields = slices.Compact(opt.MaskFields)
	opt.MaskPaths = slices.Compact(opt.MaskPaths)

	if len(opt.MaskFields) > 0 {
		opt.Processors = append(opt.Processors, MaskFieldProcessor(maskValue, opt.MaskFields...))
	}
	if len(opt.MaskPaths) > 0 {
		opt.Processors = append(opt.Processors, MaskPathProcessor(maskValue, opt.MaskPaths...))
	}
	if len(opt.IgnorePaths) > 0 {
		opt.IgnorePathsProcessor = append(opt.Processors, MaskPathProcessor(ignoreValue, opt.IgnorePaths...))
	}
	opt.Processors = append(opt.Processors, IndentProcessor)
	opt.CmpOpts = append(opt.CmpOpts, IgnoreMapEntries(opt.IgnoreFields...))

	return opt
}

type Option interface {
	isOption()
}

type Processor func(b []byte) ([]byte, error)

func (Processor) isOption() {}

type processors []Processor

func (processors) isOption() {}

func Processors(ps ...Processor) processors {
	return processors(ps)
}

type cmpOpts []cmp.Option

func (cmpOpts) isOption() {}

func CmpOpts(opts ...cmp.Option) cmpOpts {
	return cmpOpts(opts)
}

type ignoreFields []string

func (ignoreFields) isOption() {}

type ignorePaths []string

func (ignorePaths) isOption() {}

type maskFields []string

func (maskFields) isOption() {}

type maskPaths []string

func (maskPaths) isOption() {}

func IgnoreFields(fields ...string) ignoreFields {
	res := make(ignoreFields, len(fields))
	for i, field := range fields {
		res[i] = field
	}
	return res
}

func IgnorePaths(paths ...string) ignorePaths {
	res := make(ignorePaths, len(paths))
	for i, p := range paths {
		res[i] = p
	}
	return res
}

// MaskFields masks the field names. It does not take into consideration the
// path.
func MaskFields(fields ...string) maskFields {
	res := make(maskFields, len(fields))
	for i, field := range fields {
		res[i] = field
	}
	return res
}

// MaskPaths maskValue the field names. It takes into consideration the full path.
// For nested object, it may look like "$.foo.bar".
// For nested array, it may look like "$.foo[0].bar".
func MaskPaths(paths ...string) maskPaths {
	res := make(maskPaths, len(paths))
	for i, p := range paths {
		res[i] = p
	}
	return res
}

type Name string

func (n Name) isOption() {}

func IgnoreMapEntries(keys ...string) cmp.Option {
	return cmpopts.IgnoreMapEntries(func(k string, v any) bool {
		for _, key := range keys {
			if key == k {
				return true
			}
		}

		return false
	})
}
