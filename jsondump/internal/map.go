package internal

import (
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/alextanhongpin/testdump/pkg/reviver"
)

func LoadMapValues(a any, paths ...string) map[string]any {
	if len(paths) == 0 {
		return nil
	}

	res := make(map[string]any)
	_ = reviver.Walk(a, func(keys []string, v any) error {
		k := strings.Join(keys, ".")
		if slices.Contains(paths, k) {
			res[k] = v
		}

		// Terminate early
		if len(res) == len(paths) {
			return io.EOF
		}
		return nil
	})

	return res
}

func DeleteMapValues(a any, paths ...string) (any, error) {
	if len(paths) == 0 {
		return a, nil
	}

	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	var c any
	err = reviver.Unmarshal(b, &c, func(keys []string, v any) (any, error) {
		k := strings.Join(keys, ".")
		if slices.Contains(paths, k) {
			return nil, nil
		}

		return v, nil
	})
	return c, err
}
