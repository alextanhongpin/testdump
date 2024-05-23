package internal

import (
	"encoding/json"
	"slices"

	"github.com/alextanhongpin/testdump/pkg/reviver"
)

// MaskFields masks the field names. It does not take into
// consideration the path.
func MaskFields(mask string, fields []string) func([]byte) ([]byte, error) {
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

func MaskPaths(mask string, paths []string) func([]byte) ([]byte, error) {
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
