package httpdump

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alextanhongpin/testdump/httpdump/internal"
	"github.com/alextanhongpin/testdump/pkg/file"
)

type options struct {
	transformers []Transformer
	cmpOpt       CompareOption
	// The environment name to update the file.
	env  string
	file string
	// Indent the payload if the type is json. This changes the header's
	// content-length.
	indentJSON bool
	colors     bool
	body       bool
}

// newOptions is a function that takes a variadic list of options and returns a new options instance with these options.
func newOptions() *options {
	return &options{
		indentJSON: true,
		colors:     true,
	}
}

func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func (o *options) overwrite() bool {
	t, _ := strconv.ParseBool(os.Getenv(o.env))
	return t
}

func (o *options) encoder() *encoder {
	return &encoder{
		marshalFns: o.transformers,
		indentJSON: o.indentJSON,
	}
}

func (o *options) comparer() *comparer {
	return &comparer{
		cmpOpt: o.cmpOpt,
		colors: o.colors,
	}
}

func (o *options) newReadWriteCloser(name string) (io.ReadWriteCloser, error) {
	path := filepath.Join("testdata", fmt.Sprintf("%s.http", filepath.Join(name, o.file)))

	return file.New(path, o.overwrite())
}

func (o *options) newOutput(name, ext string) (io.ReadWriteCloser, error) {
	path := filepath.Join("testdata", fmt.Sprintf("%s%s", filepath.Join(name, o.file), ext))

	return file.New(path, true)
}

// Option is a type that defines a function that takes an options instance and modifies it.
type Option func(o *options)

// IndentJSON is a function that takes a boolean and returns an options that sets the indentJSON field of an options instance to the given boolean.
func IndentJSON(indent bool) Option {
	return func(o *options) {
		o.indentJSON = indent
	}
}

// CmpOpt is a function that takes a CompareOption and returns an options that merges the CompareOption with the cmpOpt field of an options instance.
func CmpOpt(opt CompareOption) Option {
	return func(o *options) {
		o.cmpOpt = o.cmpOpt.Merge(opt)
	}
}

// Env is a function that takes a string and returns an options that sets the env field of an options instance to the given string.
func Env(env string) Option {
	return func(o *options) {
		o.env = env
	}
}

// File allows setting the file name to write the output to.
func File(file string) Option {
	return func(o *options) {
		o.file = file
	}
}

// Body is a function that takes a bool and returns an options that writes the body to the file if true.
func Body(body bool) Option {
	return func(o *options) {
		o.body = body
	}
}

// Colors is a function that takes a boolean and returns an options that sets the colors field of an options instance to the given boolean.
func Colors(colors bool) Option {
	return func(o *options) {
		o.colors = colors
	}
}

// IgnoreRequestHeaders is a function that takes a variadic list of strings and returns an options that appends the strings to the Request.Header field of the cmpOpt field of an options instance.
func IgnoreRequestHeaders(val ...string) Option {
	return func(o *options) {
		o.cmpOpt.Request.Header = append(o.cmpOpt.Request.Header, internal.IgnoreMapEntries(val...))
	}
}

// IgnoreResponseHeaders is similar to IgnoreRequestHeaders, but it appends the strings to the Response.Header field of the cmpOpt field of an options instance.
func IgnoreResponseHeaders(val ...string) Option {
	return func(o *options) {
		o.cmpOpt.Response.Header = append(o.cmpOpt.Response.Header, internal.IgnoreMapEntries(val...))
	}
}

// IgnoreRequestFields is similar to IgnoreRequestHeaders, but it appends the strings to the Request.Body field of the cmpOpt field of an options instance.
func IgnoreRequestFields(val ...string) Option {
	return func(o *options) {
		o.cmpOpt.Request.Body = append(o.cmpOpt.Request.Body, internal.IgnoreMapEntries(val...))
	}
}

// IgnoreResponseFields is similar to IgnoreRequestFields, but it appends the strings to the Response.Body field of the cmpOpt field of an options instance.
func IgnoreResponseFields(val ...string) Option {
	return func(o *options) {
		o.cmpOpt.Response.Body = append(o.cmpOpt.Response.Body, internal.IgnoreMapEntries(val...))
	}
}

// Transformers is a function that takes a variadic list of transformers and returns an options that appends the transformers to the transformers field of an options instance.
func Transformers(ts ...Transformer) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, ts...)
	}
}

// MaskRequestHeaders is a function that takes a mask string and a variadic list of fields and returns an options that appends a new transformer to the transformers field of an options instance.
// The new transformer masks the request headers specified by the fields with the mask string.
func MaskRequestHeaders(mask string, fields ...string) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, maskRequestHeaders(mask, fields...))
	}
}

// MaskResponseHeaders is similar to MaskRequestHeaders, but the new transformer masks the response headers.
func MaskResponseHeaders(mask string, fields ...string) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, maskResponseHeaders(mask, fields...))
	}
}

// MaskRequestFields is similar to MaskRequestHeaders, but the new transformer masks the request fields.
func MaskRequestFields(mask string, fields ...string) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, maskRequestFields(mask, fields...))
	}
}

// MaskResponseFields is similar to MaskRequestFields, but the new transformer masks the response fields.
func MaskResponseFields(mask string, fields ...string) Option {
	return func(o *options) {
		o.transformers = append(o.transformers, maskResponseFields(mask, fields...))
	}
}
