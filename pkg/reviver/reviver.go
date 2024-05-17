package reviver

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ReviverFunc func(p, k string, v any) (any, error)

func Unmarshal(b []byte, fn ReviverFunc) (any, error) {
	var a any
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}

	var recurse func(string, any) error
	recurse = func(p string, a any) error {
		switch m := a.(type) {
		case map[string]any:
			for k, v := range m {
				p := fmt.Sprintf("%s.%s", p, k)
				o, err := fn(p, k, v)
				if err != nil {
					return err
				}
				m[k] = o
				recurse(p, o)
			}
		case []any:
			res := make([]any, len(m))
			for i, a := range m {
				p := fmt.Sprintf("%s[%d]", p, i)
				var key string
				parts := strings.Split(p, ".")
				key = parts[len(parts)-1]
				o, err := fn(p, key, a)
				if err != nil {
					res[i] = o
				}
				recurse(key, a)
			}
		default:
			// Skip primitives.
		}

		return nil
	}

	if err := recurse("$", a); err != nil {
		return nil, err
	}
	return a, nil
}
