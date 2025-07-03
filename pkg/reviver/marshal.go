package reviver

import "encoding/json"

// Marshal takes an arbitrary value and applies a function to each key-value pair.
func Marshal(a any, fn func(k []string, v any) (any, error)) ([]byte, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	var c any
	if err := Unmarshal(b, &c, fn); err != nil {
		return nil, err
	}

	return json.Marshal(c)
}
