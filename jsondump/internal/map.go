package internal

import (
	"encoding/json"
	"io"
	"slices"

	"github.com/alextanhongpin/testdump/pkg/reviver"
)

func GetMapValues(a any, paths ...string) map[string]any {
	if len(paths) == 0 {
		return nil
	}

	res := make(map[string]any)
	_ = reviver.Walk(a, func(k string, v any) error {
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

	set := make(map[string]struct{})
	for _, p := range paths {
		set[p] = struct{}{}
	}
	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	var c any
	err = reviver.Unmarshal(b, &c, func(k string, v any) (any, error) {
		if _, ok := set[k]; ok {
			delete(set, k)
			return nil, nil
		}

		return v, nil
	})
	return c, err
}
