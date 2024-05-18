package json_test

import (
	"bytes"
	"testing"
	"time"

	jsondump "github.com/alextanhongpin/dump/json"
)

func TestDump(t *testing.T) {
	type user struct {
		name string
		age  int
	}
	type User struct {
		Name string
		Age  int
	}
	testcases := []struct {
		name string
		data any
	}{
		{"dump int", 1},
		{"dump float", 0.35},
		{"dump map", map[string]any{"key": "value"}},
		{"dump slice", []int{1, 2, 3}},
		{"dump private struct", user{name: "john", age: 13}},
		{"dump public struct", User{Name: "john", Age: 13}},
		{"dump nil", nil},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			jsondump.Dump(t, tc.data)
		})
	}
}

func TestIgnoreFields(t *testing.T) {
	type Banner struct {
		ExpiresIn time.Time `json:"expiresIn"`
	}
	type MultiBanner struct {
		Banner1 Banner `json:"banner1"`
		Banner2 Banner `json:"banner2"`
	}

	banner := MultiBanner{
		Banner1: Banner{
			ExpiresIn: time.Now().Add(1 * time.Hour),
		},
		Banner2: Banner{
			ExpiresIn: time.Now().Add(1 * time.Hour),
		},
	}

	// NOTE: the name is the json tag name.
	jsondump.Dump(t, banner, jsondump.IgnoreFields("expiresIn"))
}

func TestIgnorePaths(t *testing.T) {
	type Banner struct {
		ExpiresIn time.Time `json:"expiresIn"`
	}
	type MultiBanner struct {
		Banner1 Banner `json:"banner1"`
		Banner2 Banner `json:"banner2"`
	}

	banner := MultiBanner{
		Banner1: Banner{
			ExpiresIn: time.Now().Add(1 * time.Hour),
		},
		Banner2: Banner{
			ExpiresIn: time.Now().Add(1 * time.Hour),
		},
	}

	// NOTE: the name is the json tag name.
	jsondump.Dump(t, banner,
		jsondump.IgnorePaths("$.banner1.expiresIn", "$.banner2.expiresIn"),
	)
}

func TestIgnoreFromStructTag(t *testing.T) {
	type Banner struct {
		ExpiresIn time.Time `json:"expiresIn" cmp:"-"`
	}
	type MultiBanner struct {
		Banner1 Banner `json:"banner1"`
		Banner2 Banner `json:"banner2"`
	}

	banner := MultiBanner{
		Banner1: Banner{
			ExpiresIn: time.Now().Add(1 * time.Hour),
		},
		Banner2: Banner{
			ExpiresIn: time.Now().Add(1 * time.Hour),
		},
	}

	// NOTE: the name is the json tag name.
	jsondump.Dump(t, banner)
}

func TestMaskFields(t *testing.T) {
	type Account struct {
		Type  string `json:"type"`
		Email string `json:"email"`
	}

	type Accounts struct {
		Email  Account `json:"email"`
		Google Account `json:"google"`
	}

	accounts := Accounts{
		Email: Account{
			Type:  "email",
			Email: "john.appleseed@mail.com",
		},
		Google: Account{
			Type:  "google",
			Email: "john.appleseed@gmail.com",
		},
	}

	jsondump.Dump(t, accounts, jsondump.MaskFields("email"))
}

func TestMaskFieldsFromStructTag(t *testing.T) {
	type Account struct {
		Type  string `json:"type"`
		Email string `json:"email" mask:"true"`
	}

	type Accounts struct {
		Email  Account `json:"email"`
		Google Account `json:"google"`
	}

	accounts := Accounts{
		Email: Account{
			Type:  "email",
			Email: "john.appleseed@mail.com",
		},
		Google: Account{
			Type:  "google",
			Email: "john.appleseed@gmail.com",
		},
	}

	jsondump.Dump(t, accounts)
}

func TestMaskPaths(t *testing.T) {
	type Account struct {
		Type  string `json:"type"`
		Email string `json:"email"`
	}

	type Accounts struct {
		Email  Account `json:"email"`
		Google Account `json:"google"`
	}

	accounts := Accounts{
		Email: Account{
			Type:  "email",
			Email: "john.appleseed@mail.com",
		},
		Google: Account{
			Type:  "google",
			Email: "john.appleseed@gmail.com",
		},
	}

	jsondump.Dump(t, accounts, jsondump.MaskPaths("$.email.email"))
}

func TestCustomProcessor(t *testing.T) {
	jsondump.Dump(t, map[string]any{
		"name": "John",
	}, jsondump.Processor(func(b []byte) ([]byte, error) {
		return bytes.ToUpper(b), nil
	}))
}

func TestMultipleProcessors(t *testing.T) {
	jsondump.Dump(t, map[string]any{
		"name": "John",
	}, jsondump.Processors(
		func(b []byte) ([]byte, error) {
			return bytes.ToLower(b), nil
		},
		func(b []byte) ([]byte, error) {
			return bytes.ToTitle(b), nil
		},
	))
}

func TestCustomName(t *testing.T) {
	// If you need to write multiple files for the same test, you can customize
	// the name to avoid overwriting the existing file.
	john := map[string]any{
		"name": "John",
	}
	jane := map[string]any{
		"name": "Jane",
	}

	jsondump.Dump(t, john, jsondump.Name("john"))
	jsondump.Dump(t, jane, jsondump.Name("jane"))
}
