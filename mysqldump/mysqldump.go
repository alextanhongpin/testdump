package mysqldump

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alextanhongpin/dump/mysqldump/internal"
)

type Option interface {
	isOption()
}

type File string

func (n File) isOption() {}

type option struct {
	File string
}

func newOption(opts ...Option) option {
	var o option
	for _, opt := range opts {
		switch v := opt.(type) {
		case File:
			o.File = string(v)
		}
	}
	return o
}

func Dump(t *testing.T, received *SQL, opts ...Option) {
	t.Helper()
	if err := dump(t, received, opts...); err != nil {
		t.Error(err)
	}
}

func dump(t *testing.T, received *SQL, opts ...Option) error {
	opt := newOption(opts...)
	receivedBytes, err := Write(received)
	if err != nil {
		return err
	}

	file := filepath.Join("testdata", fmt.Sprintf("%s.sql", filepath.Join(t.Name(), opt.File)))
	overwrite := false
	written, err := internal.WriteFile(file, receivedBytes, overwrite)
	if err != nil {
		return err
	}

	if written {
		return nil
	}

	snapshotBytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	snapshot, err := Read(snapshotBytes)
	if err != nil {
		return err
	}

	if err := Compare(snapshot, received); err != nil {
		if opt.File != "" {
			return fmt.Errorf("%s: %w", opt.File, err)
		}

		return err
	}

	return nil
}

func toMap(s []any) (any, error) {
	m := make(map[string]any)
	for k, v := range s {
		m[fmt.Sprintf(":v%d", k+1)] = v
	}

	// Marshal/unmarshal to avoid type issues such as
	// int/float.
	// In JSON, there's only float.
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var a any
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}

	return a, nil
}
