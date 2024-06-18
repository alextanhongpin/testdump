package internal

import (
	"fmt"
	"reflect"
	"strings"
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

func IterStructFields(a any, fn func(k string, f reflect.StructField, v reflect.Value)) error {
	var recurse func(k string, t reflect.StructField, v reflect.Value) error
	recurse = func(k string, t reflect.StructField, v reflect.Value) error {
		if v.Kind() == reflect.Ptr {
			if next := v.Elem(); next.Kind() == reflect.Struct && v.IsNil() {
				return nil
			} else {
				v = next
			}
		}
		if v.Kind() == reflect.Invalid {
			return nil
		}

		switch reflect.Indirect(v).Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				fv := v.Index(i)
				name := fmt.Sprintf("%s[%d]", k, i)
				if err := recurse(name, reflect.StructField{}, fv); err != nil {
					return err
				}
			}
		case reflect.Map:
			iter := v.MapRange()
			for iter.Next() {
				key := iter.Key()
				name := fmt.Sprintf("%s.%s", k, key)
				if err := recurse(name, reflect.StructField{}, iter.Value()); err != nil {
					return err
				}
			}

			return nil
		case reflect.Struct:
			for _, f := range reflect.VisibleFields(v.Type()) {
				// Could be external packages like time.Time.
				// We don't want to recurse into these.
				if f.PkgPath != "" {
					fn(k, t, v)
					continue
				}

				fv := v.FieldByName(f.Name)
				name := fmt.Sprintf("%s.%s", k, Or(jsonTag(f), f.Name))
				if err := recurse(name, f, fv); err != nil {
					return err
				}
			}
		default:
			fn(k, t, v)

		}
		return nil
	}
	return recurse("$", reflect.StructField{}, reflect.ValueOf(a))
}

func jsonTag(f reflect.StructField) string {
	fields := f.Tag.Get("json")
	paths := strings.Split(fields, ",")
	return paths[0]
}

func Or[T comparable](vs ...T) T {
	var z T
	for _, v := range vs {
		if v != z {
			return v
		}
	}
	return z
}
