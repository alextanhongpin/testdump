package internal

import (
	"slices"
	"strings"
)

type fieldFunc = func([]string, any) (any, error)

func MaskFieldsFunc(mask string, fields []string) fieldFunc {
	return func(keys []string, val any) (any, error) {
		if len(keys) == 0 {
			return val, nil
		}

		field := keys[len(keys)-1]
		if slices.Contains(fields, field) {
			// Allows masking string values only for now.
			if _, ok := val.(string); ok {
				return mask, nil
			}
		}

		return val, nil
	}
}

func MaskPathsFunc(mask string, paths []string) fieldFunc {
	return func(keys []string, val any) (any, error) {
		path := strings.Join(keys, ".")
		if slices.Contains(paths, path) {
			// Allows masking string values only for now.
			_, ok := val.(string)
			if ok {
				return mask, nil
			}
		}

		return val, nil
	}
}
