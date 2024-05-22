package jsondump_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/alextanhongpin/dump/jsondump"
	"github.com/alextanhongpin/dump/pkg/cuetest"
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
	}, jsondump.Processor(
		func(b []byte) ([]byte, error) {
			return bytes.ToLower(b), nil
		}),
		jsondump.Processor(
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

	jsondump.Dump(t, john, jsondump.File("john"))
	jsondump.Dump(t, jane, jsondump.File("jane"))
}

func TestCUESchema(t *testing.T) {
	type User struct {
		Name     string   `json:"name"`
		Age      int      `json:"age"`
		Birthday string   `json:"birthday"`
		Hobbies  []string `json:"hobbies"`
		ImageURL string   `json:"imageURL"`
	}

	u := User{
		Name:     "John",
		Age:      13,
		Birthday: time.Now().Format(time.DateOnly),
		Hobbies:  []string{"reading"},
		ImageURL: "https://example.com/image.jpg",
	}

	c := &cuetest.Validator{
		Schemas: []string{`package test
// https://cuelang.org/docs/howto/

import "list"
import "time"
import "strings"

let url = =~ "^https://(.+)"

#User: close({
    name!: string & strings.MinRunes(2) & strings.MaxRunes(8)
    age!: >= 13
    hobbies!: [...string] & list.MinItems(1) & list.MaxItems(1)
    birthday!: string & time.Format("2006-01-02")
    imageURL: string & url
})

#User`, // Set the root to #User.
		},
	}

	jsondump.Dump(t, u, jsondump.Processor(func(b []byte) ([]byte, error) {
		return b, c.Validate(b)
	}))
}

func TestCUESchemaField(t *testing.T) {
	type User struct {
		Name     string   `json:"name"`
		Age      int      `json:"age"`
		Birthday string   `json:"birthday"`
		Hobbies  []string `json:"hobbies"`
		ImageURL string   `json:"imageURL"`
	}

	u := User{
		Name:     "John",
		Age:      13,
		Birthday: time.Now().Format(time.DateOnly),
		Hobbies:  []string{"reading"},
		ImageURL: "https://example.com/image.jpg",
	}

	c := &cuetest.Validator{
		// We don't need to do full validation.
		// This is just a simpler way to do assertions.
		Schemas: []string{
			`age!: >= 13`,
			`import "strings"
name!: string & strings.MinRunes(1)`,
		},
	}

	jsondump.Dump(t, u, jsondump.Processor(func(b []byte) ([]byte, error) {
		return b, c.Validate(b)
	}))
}

func TestCUESchemaPath(t *testing.T) {
	type User struct {
		Name     string   `json:"name"`
		Age      int      `json:"age"`
		Birthday string   `json:"birthday"`
		Hobbies  []string `json:"hobbies"`
		ImageURL string   `json:"imageURL"`
	}

	u := User{
		Name:     "John",
		Age:      13,
		Birthday: time.Now().Format(time.DateOnly),
		Hobbies:  []string{"reading"},
		ImageURL: "https://example.com/image.jpg",
	}

	c := &cuetest.Validator{
		SchemaPaths: []string{
			"./testdata/user.cue",
			"./testdata/test.cue",
		},
	}

	jsondump.Dump(t, u, jsondump.Processor(func(b []byte) ([]byte, error) {
		return b, c.Validate(b)
	}))
}
