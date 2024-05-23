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

// newOption is a function that takes a variadic list of options and returns a new option instance with these options.
func newOption(opts ...Option) *option {
	o := new(option)
	o.indentJSON = true
	o.colors = true

	// Apply each option to the new option instance.
	for _, opt := range opts {
		opt(o)
	}

	return o
}

// Option is a type that defines a function that takes an option instance and modifies it.
type Option func(o *option)

// IndentJSON is a function that takes a boolean and returns an option that sets the indentJSON field of an option instance to the given boolean.
func IndentJSON(indent bool) Option {
	return func(o *option) {
		o.indentJSON = indent
	}
}

// CmpOpt is a function that takes a CompareOption and returns an option that merges the CompareOption with the cmpOpt field of an option instance.
func CmpOpt(opt CompareOption) Option {
	return func(o *option) {
		o.cmpOpt = o.cmpOpt.Merge(opt)
	}
}

// Env is a function that takes a string and returns an option that sets the env field of an option instance to the given string.
func Env(env string) Option {
	return func(o *option) {
		o.env = env
	}
}

// Colors is a function that takes a boolean and returns an option that sets the colors field of an option instance to the given boolean.
func Colors(colors bool) Option {
	return func(o *option) {
		o.colors = colors
	}
}

// IgnoreRequestHeaders is a function that takes a variadic list of strings and returns an option that appends the strings to the Request.Header field of the cmpOpt field of an option instance.
func IgnoreRequestHeaders(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Request.Header = append(o.cmpOpt.Request.Header, internal.IgnoreMapEntries(val...))
	}
}

// IgnoreResponseHeaders is similar to IgnoreRequestHeaders, but it appends the strings to the Response.Header field of the cmpOpt field of an option instance.
func IgnoreResponseHeaders(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Response.Header = append(o.cmpOpt.Response.Header, internal.IgnoreMapEntries(val...))
	}
}

// IgnoreRequestFields is similar to IgnoreRequestHeaders, but it appends the strings to the Request.Body field of the cmpOpt field of an option instance.
func IgnoreRequestFields(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Request.Body = append(o.cmpOpt.Request.Body, internal.IgnoreMapEntries(val...))
	}
}

// IgnoreResponseFields is similar to IgnoreRequestFields, but it appends the strings to the Response.Body field of the cmpOpt field of an option instance.
func IgnoreResponseFields(val ...string) Option {
	return func(o *option) {
		o.cmpOpt.Response.Body = append(o.cmpOpt.Response.Body, internal.IgnoreMapEntries(val...))
	}
}

// Transformers is a function that takes a variadic list of transformers and returns an option that appends the transformers to the transformers field of an option instance.
func Transformers(ts ...Transformer) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, ts...)
	}
}

// MaskRequestHeaders is a function that takes a mask string and a variadic list of fields and returns an option that appends a new transformer to the transformers field of an option instance.
// The new transformer masks the request headers specified by the fields with the mask string.
func MaskRequestHeaders(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskRequestHeaders(mask, fields...))
	}
}

// MaskResponseHeaders is similar to MaskRequestHeaders, but the new transformer masks the response headers.
func MaskResponseHeaders(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskResponseHeaders(mask, fields...))
	}
}

// MaskRequestFields is similar to MaskRequestHeaders, but the new transformer masks the request fields.
func MaskRequestFields(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskRequestFields(mask, fields...))
	}
}

// MaskResponseFields is similar to MaskRequestFields, but the new transformer masks the response fields.
func MaskResponseFields(mask string, fields ...string) Option {
	return func(o *option) {
		o.transformers = append(o.transformers, maskResponseFields(mask, fields...))
	}
}
