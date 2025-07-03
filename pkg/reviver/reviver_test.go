package reviver_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/alextanhongpin/testdump/pkg/reviver"
	"github.com/google/go-cmp/cmp"
)

func TestReviver(t *testing.T) {
	type Account struct {
		Type string `json:"type"`
	}

	type User struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Married bool     `json:"married"`
		Hobbies []string `json:"hobbies"`
		Info    struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"info"`
		Accounts []Account `json:"accounts"`
	}

	o := User{
		Name:    "John",
		Age:     13,
		Married: false,
		Hobbies: []string{"swimming", "jogging"},
		Info: struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}{
			FirstName: "John",
			LastName:  "Appleseed",
		},
		Accounts: []Account{
			{Type: "email"},
			{Type: "facebook"},
		},
	}

	b, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}

	// We unmarshal/marshal back to map[string]any to fix the ordering issue.
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	b, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("iterates through all key/value pairs", func(t *testing.T) {
		var n map[string]any
		err := reviver.Unmarshal([]byte(b), &n, func(k []string, v any) (any, error) {
			t.Log(k, v)
			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unmarshal to map[string]any", func(t *testing.T) {
		var n map[string]any
		err := reviver.Unmarshal([]byte(b), &n, func(k []string, v any) (any, error) {
			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, n, true)
	})

	t.Run("unmarshal to struct", func(t *testing.T) {
		var u User
		err := reviver.Unmarshal([]byte(b), &u, func(k []string, v any) (any, error) {
			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, u, true)
	})

	t.Run("override a field string", func(t *testing.T) {
		var u User
		err := reviver.Unmarshal([]byte(b), &u, func(k []string, v any) (any, error) {
			if tail(k) == "name" {
				return "Jane", nil
			}

			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, u, false)
	})

	t.Run("override a field slice", func(t *testing.T) {
		var u User
		err := reviver.Unmarshal([]byte(b), &u, func(k []string, v any) (any, error) {
			if tail(k) == "hobbies" {
				return []string{"programming"}, nil
			}

			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, u, false)
	})

	t.Run("override a field map", func(t *testing.T) {
		// Need to use map[string]any, cause User does not have the new field.
		var u map[string]any
		err := reviver.Unmarshal([]byte(b), &u, func(k []string, v any) (any, error) {
			if tail(k) == "info" {
				return map[string]any{
					"first_name":  "Jane",
					"middle_name": `"Better"`,
				}, nil
			}

			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, u, false)
	})

	t.Run("override a slice item", func(t *testing.T) {
		var u User
		err := reviver.Unmarshal([]byte(b), &u, func(k []string, v any) (any, error) {
			if tail(k) == "hobbies[0]" {
				return "programming", nil
			}

			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, u, false)
	})
}

func TestWalk(t *testing.T) {
	t.Run("simple object traversal", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
			"city": "New York",
		}

		var visitedPaths []string
		var visitedValues []any

		err := reviver.Walk(data, func(path []string, value any) error {
			visitedPaths = append(visitedPaths, joinPath(path))
			visitedValues = append(visitedValues, value)
			t.Logf("Path: %v, Value: %v", path, value)
			return nil
		})

		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Should visit the root object and all leaf values
		expectedPaths := []string{"", "age", "city", "name"}
		if len(visitedPaths) != len(expectedPaths) {
			t.Errorf("Expected %d visits, got %d", len(expectedPaths), len(visitedPaths))
		}
	})

	t.Run("nested object traversal", func(t *testing.T) {
		jsonData := `{
			"user": {
				"name": "Alice",
				"profile": {
					"email": "alice@example.com",
					"settings": {
						"theme": "dark"
					}
				}
			}
		}`

		var data any
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			t.Fatal(err)
		}

		var pathLog []string
		var depthLog []int

		err := reviver.Walk(data, func(path []string, value any) error {
			pathLog = append(pathLog, joinPath(path))
			depthLog = append(depthLog, len(path))
			t.Logf("Depth %d: Path=%s, Type=%T", len(path), joinPath(path), value)
			return nil
		})

		if err != nil {
			t.Fatalf("Walk failed: %v", err)
		}

		// Check that we visit all levels
		maxDepth := 0
		for _, depth := range depthLog {
			if depth > maxDepth {
				maxDepth = depth
			}
		}

		if maxDepth < 3 {
			t.Errorf("Expected to traverse at least 3 levels deep, got %d", maxDepth)
		}

		t.Logf("Visited %d nodes with max depth %d", len(pathLog), maxDepth)
	})

	t.Run("error handling", func(t *testing.T) {
		data := map[string]any{
			"good": "value",
			"bad":  "trigger error",
		}

		err := reviver.Walk(data, func(path []string, value any) error {
			if str, ok := value.(string); ok && str == "trigger error" {
				t.Logf("Triggering error at path: %v", path)
				return fmt.Errorf("intentional error")
			}
			return nil
		})

		if err == nil {
			t.Fatal("Expected error but got none")
		}

		t.Logf("Got expected error: %v", err)
	})
}

// Helper functions for tests

func joinPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	result := path[0]
	for _, segment := range path[1:] {
		if containsArrayIndex(segment) {
			result += segment
		} else {
			result += "." + segment
		}
	}
	return result
}

func containsArrayIndex(segment string) bool {
	return len(segment) > 0 && segment[len(segment)-1] == ']'
}

func diffJSON(t *testing.T, a, b any, errorOnDiff bool) {
	t.Helper()

	// process marshal/unmarshal the struct to a map to remove the ordering of
	// the fields which affects comparison.
	process := func(a any) (any, error) {
		if _, ok := a.(map[string]any); ok {
			return a, nil
		}

		b, err := json.Marshal(a)
		if err != nil {
			return nil, err
		}

		var c any
		if err := json.Unmarshal(b, &c); err != nil {
			return nil, err
		}

		return c, nil
	}

	x, err := process(a)
	if err != nil {
		t.Fatal(err)
	}

	y, err := process(b)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(x, y); diff != "" {
		if errorOnDiff {
			t.Error(diff)
		} else {
			t.Log(diff)
		}
	}
}

func tail(p []string) string {
	if len(p) == 0 {
		return ""
	}

	return p[len(p)-1]
}
