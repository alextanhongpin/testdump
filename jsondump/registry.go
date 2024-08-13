package jsondump

import "reflect"

type Registry struct {
	opts map[any][]Option
}

func NewRegistry() *Registry {
	return &Registry{
		opts: make(map[any][]Option),
	}
}

func (r *Registry) Register(v any, opts ...Option) {
	r.opts[indirectType(v)] = opts
}

func (r *Registry) Get(v any) []Option {
	return r.opts[indirectType(v)]
}

// Ensures that the type is not a pointer.
func indirectType(v any) reflect.Type {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}
