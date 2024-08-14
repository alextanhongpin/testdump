package yamldump_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/alextanhongpin/testdump/yamldump"
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
			yamldump.Dump(t, tc.data)
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
	yamldump.Dump(t, banner, yamldump.IgnoreFields("expiresIn"))
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
	yamldump.Dump(t, banner,
		yamldump.IgnorePaths("$.banner1.expiresIn", "$.banner2.expiresIn"),
	)
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

	// Create a consistent maskFields function.
	m := yamldump.NewMask("REDACTED")
	yamldump.Dump(t, accounts, m.MaskFields("email"))
}

func TestNew(t *testing.T) {
	type User struct {
		Password  string    `json:"password"`
		CreatedAt time.Time `json:"createdAt"`
	}

	yd := yamldump.New(
		yamldump.IgnorePaths("$.createdAt"),
		yamldump.MaskPaths("[REDACTED]", []string{"$.password"}),
	)
	yd.Dump(t, User{
		Password:  "password",
		CreatedAt: time.Now(), // Dynamic value
	})
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

	yamldump.Dump(t, accounts, yamldump.MaskPaths("[MASKED]", []string{"$.email.email"}))
}

func TestCustomTransformer(t *testing.T) {
	yamldump.Dump(t, map[string]any{
		"name": "John",
	}, yamldump.Transformers(func(b []byte) ([]byte, error) {
		return bytes.ToUpper(b), nil
	}))
}

func TestMultipleTransformers(t *testing.T) {
	yamldump.Dump(t, map[string]any{
		"name": "John",
	}, yamldump.Transformers(
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
		"age":  13,
	}
	jane := map[string]any{
		"name": "Jane",
		"age":  13,
	}

	yamldump.Dump(t, john, yamldump.File("john"))
	yamldump.Dump(t, jane, yamldump.File("jane"))
}

func TestRegistry(t *testing.T) {
	type Foo struct {
		CreatedAt time.Time
	}
	type Bar struct {
		UpdatedAt time.Time
	}

	yd := yamldump.New()
	yd.Register(&Foo{}, yamldump.IgnoreFields("CreatedAt"))
	yd.Register(Bar{}, yamldump.IgnoreFields("UpdatedAt"))

	yd.Dump(t, Foo{CreatedAt: time.Now()})
	yd.Dump(t, &Bar{UpdatedAt: time.Now()})
}
