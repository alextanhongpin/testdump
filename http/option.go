package http

import (
	"github.com/alextanhongpin/dump/http/internal"
)

type Option interface {
	isOption()
}

type option struct {
	mws        []Middleware
	req        CompareOption
	res        CompareOption
	indentJSON bool
}

func newOption(opts ...Option) *option {
	opt := new(option)
	opt.indentJSON = true
	kvs := make(map[string][]string)

	for _, o := range opts {
		switch v := o.(type) {
		case Middleware:
			opt.mws = append(opt.mws, v)
		case modifier:
			v(opt)
		case kv:
			if _, ok := kvs[v.key]; !ok {
				kvs[v.key] = v.val
			} else {
				kvs[v.key] = append(kvs[v.key], v.val...)
			}
		}
	}

	opt.req.Header = append(opt.req.Header, internal.IgnoreMapEntries(kvs["request_header"]...))
	opt.res.Header = append(opt.res.Header, internal.IgnoreMapEntries(kvs["response_header"]...))
	opt.req.Body = append(opt.req.Body, internal.IgnoreMapEntries(kvs["request_field"]...))
	opt.res.Body = append(opt.res.Body, internal.IgnoreMapEntries(kvs["response_field"]...))

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

type kv struct {
	key string
	val []string
}

func (kv) isOption() {}

func IgnoreRequestHeaders(val ...string) kv {
	return kv{
		key: "request_header",
		val: val,
	}
}

func IgnoreResponseHeaders(val ...string) kv {
	return kv{
		key: "response_header",
		val: val,
	}
}

func IgnoreRequestFields(val ...string) kv {
	return kv{
		key: "request_field",
		val: val,
	}
}

func IgnoreResponseFields(val ...string) kv {
	return kv{
		key: "response_field",
		val: val,
	}
}
