package httpdump

import (
	"github.com/alextanhongpin/dump/httpdump/internal"
)

type Option interface {
	isOption()
}

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
	opt := new(option)
	opt.indentJSON = true
	opt.colors = true

	for _, o := range opts {
		switch v := o.(type) {
		case Transformer:
			opt.transformers = append(opt.transformers, v)
		case modifier:
			v(opt)
		}
	}

	return opt
}

type modifier func(o *option)

func (modifier) isOption() {}

func IndentJSON(indent bool) modifier {
	return func(o *option) {
		o.indentJSON = indent
	}
}

func CmpOpt(opt CompareOption) modifier {
	return func(o *option) {
		o.cmpOpt = o.cmpOpt.Merge(opt)
	}
}

// Env is the name environment name to update the file.
func Env(env string) modifier {
	return func(o *option) {
		o.env = env
	}
}

func Colors(colors bool) modifier {
	return func(o *option) {
		o.colors = colors
	}
}

func IgnoreRequestHeaders(val ...string) modifier {
	return func(o *option) {
		o.cmpOpt.Request.Header = append(o.cmpOpt.Request.Header, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreResponseHeaders(val ...string) modifier {
	return func(o *option) {
		o.cmpOpt.Response.Header = append(o.cmpOpt.Response.Header, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreRequestFields(val ...string) modifier {
	return func(o *option) {
		o.cmpOpt.Request.Body = append(o.cmpOpt.Request.Body, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreResponseFields(val ...string) modifier {
	return func(o *option) {
		o.cmpOpt.Response.Body = append(o.cmpOpt.Response.Body, internal.IgnoreMapEntries(val...))
	}
}
