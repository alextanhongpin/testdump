package httpdump

import (
	"github.com/alextanhongpin/testdump/httpdump/internal"
)

type option struct {
	transformers []Transformer
	cmpOpt       CompareOption
	// The environment name to update the file.
	env string
	// Indent the payload if the type is json. This changes the header's
	// content-length.
	indentJSON bool
	colors     bool
}

func newOption(opts ...Option) *option {
	o := new(option)
	o.indentJSON = true
	o.colors = true

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type Option func(o *option)

func IndentJSON(indent bool) Option {
	return func(o *option) {
		o.indentJSON = indent
	}
}

func CmpOpt(opt CompareOption) Option {
	return func(o *option) {
		o.cmpOpt = o.cmpOpt.Merge(opt)
	}
}

// Env is the name environment name to update the file.
func Env(env string) Option {
	return func(o *option) {
		o.env = env
	}
}

func Colors(colors bool) Option {
	return func(o *option) {
		o.colors = colors
	}
}

func IgnoreRequestHeaders(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Request.Header = append(o.cmpOpt.Request.Header, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreResponseHeaders(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Response.Header = append(o.cmpOpt.Response.Header, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreRequestFields(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Request.Body = append(o.cmpOpt.Request.Body, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreResponseFields(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Response.Body = append(o.cmpOpt.Response.Body, internal.IgnoreMapEntries(val...))
	}
}

func Transformers(ts ...Transformer) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, ts...)
	}
}

func MaskRequestHeaders(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskRequestHeaders(mask, fields...))
	}
}

func MaskResponseHeaders(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskResponseHeaders(mask, fields...))
	}
}

func MaskRequestFields(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskRequestFields(mask, fields...))
	}
}

func MaskResponseFields(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskResponseFields(mask, fields...))
	}
}
