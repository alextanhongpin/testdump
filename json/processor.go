package json

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"strconv"

	"github.com/alextanhongpin/dump/json/internal"
	"github.com/alextanhongpin/dump/pkg/reviver"
)

func CueProcessor(b []byte) ([]byte, error) {

	return b, nil
}

func IndentProcessor(b []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", " "); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func MaskFieldProcessor(mask string, fields ...string) Processor {
	slices.Sort(fields)
	fields = slices.Compact(fields)

	return func(b []byte) ([]byte, error) {
		var m any
		if err := reviver.Unmarshal(b, &m, func(key string, val any) (any, error) {
			path := reviver.Base(key)
			for _, f := range fields {
				if f == path {
					// Allows masking string values only for now.
					_, ok := val.(string)
					if ok {
						return mask, nil
					}
				}
			}

			return val, nil
		}); err != nil {
			return nil, err
		}

		return json.MarshalIndent(m, "", " ")
	}
}

func MaskPathProcessor(mask string, paths ...string) Processor {
	slices.Sort(paths)
	paths = slices.Compact(paths)

	return func(b []byte) ([]byte, error) {
		var m any
		if err := reviver.Unmarshal(b, &m, func(key string, val any) (any, error) {
			for _, f := range paths {
				if f == key {
					// Allows masking string values only for now.
					_, ok := val.(string)
					if ok {
						return mask, nil
					}
				}
			}

			return val, nil
		}); err != nil {
			return nil, err
		}

		return json.MarshalIndent(m, "", " ")
	}
}

func MaskPathsFromStructTag(a any) []string {
	var maskPaths []string
	internal.IterStructFields(a, func(k string, f reflect.StructField, v reflect.Value) {
		// `mask:"true"`
		tag := f.Tag.Get("mask")
		ok, _ := strconv.ParseBool(tag)
		if ok {
			maskPaths = append(maskPaths, k)
		}
	})

	return maskPaths
}

func IgnorePathsFromStructTag(a any) []string {
	var maskPaths []string
	internal.IterStructFields(a, func(k string, f reflect.StructField, v reflect.Value) {
		// `cmp:"ignore"`
		tag := f.Tag.Get("cmp")
		if tag == "-" {
			maskPaths = append(maskPaths, k)
		}
	})

	return maskPaths
}
