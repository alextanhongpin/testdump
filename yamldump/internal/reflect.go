package internal

import (
	"reflect"
)

func TypeName(v any) string {
	// Reflect will panic if there is no
	// type detected, which is usually
	// the case for nil types.
	if v == nil {
		return "nil"
	}

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return t.String()
	}

	return t.Kind().String()
}
