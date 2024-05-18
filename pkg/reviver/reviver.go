// You can edit this code!
// Click here and start typing.
package reviver

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

const root = "$"

// ReviverFunc is a function that is called for each
// key-value pair in the JSON object.
type ReviverFunc func(k string, v any) (any, error)

// Unmarshal parses the JSON-encoded data and stores the
// result in the value pointed to by t.
func Unmarshal(b []byte, t any, fn ReviverFunc) error {
	var a any
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}

	var recurse func(string, any) (any, error)
	recurse = func(p string, a any) (any, error) {
		switch m := a.(type) {
		case map[string]any:
			if v, err := fn(p, m); err != nil {
				return nil, err
			} else if !reflect.DeepEqual(v, m) {
				return v, nil
			}

			for k, v := range m {
				o, err := recurse(fmt.Sprintf("%s.%s", p, k), v)
				if err != nil {
					return nil, err
				}
				m[k] = o
			}
			return m, nil
		case []any:
			if v, err := fn(p, m); err != nil {
				return nil, err
			} else if !reflect.DeepEqual(v, m) {
				return v, nil
			}

			res := make([]any, len(m))
			for i, a := range m {
				o, err := recurse(fmt.Sprintf("%s[%d]", p, i), a)
				if err != nil {
					return nil, err
				}
				res[i] = o
			}

			return res, nil
		default:
			return fn(p, a)
		}
	}

	o, err := recurse(root, a)
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

// Base returns the base name of the path.
func Base(k string) string {
	parts := strings.Split(k, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
