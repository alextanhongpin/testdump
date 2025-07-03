// You can edit this code!
// Click here and start typing.
package reviver

import (
	"encoding/json"
	"fmt"
)

// Unmarshal parses the JSON-encoded data and stores the
// result in the value pointed to by t.
func Unmarshal(b []byte, t any, fn func(k []string, v any) (any, error)) error {
	var a any
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}

	var recurse func([]string, any) (any, error)
	recurse = func(p []string, a any) (any, error) {
		switch m := a.(type) {
		case map[string]any:
			for k, v := range m {
				o, err := recurse(append(p, k), v)
				if err != nil {
					return nil, err
				}
				m[k] = o
			}

			return fn(p, m)
		case []any:
			h, t := pop(p)
			res := make([]any, len(m))
			for i, a := range m {
				o, err := recurse(append(h, fmt.Sprintf("%s[%d]", t, i)), a)
				if err != nil {
					return nil, err
				}
				res[i] = o
			}

			return fn(p, res)
		default:
			return fn(p, a)
		}
	}

	o, err := recurse(nil, a)
	if err != nil {
		return err
	}

	// Reduce unnecessary marshal/unmarshall-ing.
	if v, ok := t.(*map[string]any); ok {
		*v = o.(map[string]any)
		return nil
	}

	b, err = json.Marshal(o)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, t)
}

type WalkFunc = func(k []string, v any) error

func Walk(a any, fn WalkFunc) error {
	var walk func([]string, any) error
	walk = func(p []string, a any) error {
		switch m := a.(type) {
		case map[string]any:
			if err := fn(p, m); err != nil {
				return err
			}

			for k, v := range m {
				if err := walk(append(p, k), v); err != nil {
					return err
				}
			}
			return nil
		case []any:
			if err := fn(p, m); err != nil {
				return err
			}

			p, tail := pop(p)
			for i, a := range m {
				if err := walk(append(p, fmt.Sprintf("%s[%d]", tail, i)), a); err != nil {
					return err
				}
			}

			return nil
		default:
			return fn(p, a)
		}
	}

	return walk(nil, a)
}

func pop(p []string) ([]string, string) {
	if len(p) == 0 {
		return nil, ""
	}

	// Create a copy to avoid sharing the underlying array
	result := make([]string, len(p)-1)
	copy(result, p[:len(p)-1])
	return result, p[len(p)-1]
}
