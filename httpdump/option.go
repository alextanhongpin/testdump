package httpdump

import (
	"github.com/alextanhongpin/dump/httpdump/internal"
)

type Option interface {
	isOption()
}

type option struct {
	middlewares []Middleware
	req         CompareOption
	res         CompareOption
	// The environment name to update the file.
	env string
	// Indent the payload if the type is json. This changes the header's
	// content-length.
	indentJSON bool
}

func newOption(opts ...Option) *option {
	opt := new(option)
	opt.indentJSON = true

	for _, o := range opts {
		switch v := o.(type) {
		case Middleware:
			opt.middlewares = append(opt.middlewares, v)
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

func RequestCmpOpt(opt CompareOption) modifier {
	return func(o *option) {
		o.req = o.req.Merge(opt)
	}
}

func ResponseCmpOpt(opt CompareOption) modifier {
	return func(o *option) {
		o.res = o.res.Merge(opt)
	}
}

// Env is the name environment name to update the file.
func Env(env string) modifier {
	return func(o *option) {
		o.env = env
	}
}

func IgnoreRequestHeaders(val ...string) modifier {
	return func(o *option) {
		o.req.Header = append(o.req.Header, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreResponseHeaders(val ...string) modifier {
	return func(o *option) {
		o.res.Header = append(o.res.Header, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreRequestFields(val ...string) modifier {
	return func(o *option) {
		o.req.Body = append(o.req.Body, internal.IgnoreMapEntries(val...))
	}
}

func IgnoreResponseFields(val ...string) modifier {
	return func(o *option) {
		o.res.Body = append(o.res.Body, internal.IgnoreMapEntries(val...))
	}
}
