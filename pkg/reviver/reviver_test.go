package reviver_test

import (
	"encoding/json"
	"testing"

	"github.com/alextanhongpin/dump/pkg/reviver"
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
		err := reviver.Unmarshal([]byte(b), &n, func(k string, v any) (any, error) {
			t.Log(k, v)
			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unmarshal to map[string]any", func(t *testing.T) {
		var n map[string]any
		err := reviver.Unmarshal([]byte(b), &n, func(k string, v any) (any, error) {
			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, n, true)
	})

	t.Run("unmarshal to struct", func(t *testing.T) {
		var u User
		err := reviver.Unmarshal([]byte(b), &u, func(k string, v any) (any, error) {
			return v, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		diffJSON(t, m, u, true)
	})

	t.Run("override a field string", func(t *testing.T) {
		var u User
		err := reviver.Unmarshal([]byte(b), &u, func(k string, v any) (any, error) {
			if reviver.Base(k) == "name" {
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
		err := reviver.Unmarshal([]byte(b), &u, func(k string, v any) (any, error) {
			if reviver.Base(k) == "hobbies" {
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
		err := reviver.Unmarshal([]byte(b), &u, func(k string, v any) (any, error) {
			if reviver.Base(k) == "info" {
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
		err := reviver.Unmarshal([]byte(b), &u, func(k string, v any) (any, error) {
			if reviver.Base(k) == "hobbies[0]" {
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
