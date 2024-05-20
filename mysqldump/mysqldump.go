package mysqldump

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alextanhongpin/dump/mysqldump/internal"
	"github.com/alextanhongpin/dump/pkg/diff"
	"vitess.io/vitess/go/vt/sqlparser"
)

type Option interface {
	isOption()
}

type Name string

func (n Name) isOption() {}

type option struct {
	Name string
}

func newOption(opts ...Option) option {
	var o option
	for _, opt := range opts {
		switch v := opt.(type) {
		case Name:
			o.Name = string(v)
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

	file := filepath.Join("testdata", fmt.Sprintf("%s.sql", filepath.Join(t.Name(), opt.Name)))
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

	x := snapshot
	y := received
	ok, err := Compare(x.Query, y.Query)
	if err != nil {
		return err
	}

	if !ok {
		label := "Query"
		if opt.Name != "" {
			label = fmt.Sprintf("%s %s", label, opt.Name)
		}
		return fmt.Errorf("%s: %w", label, diff.ANSI(x.Query, y.Query))
	}

	{
		lhs, err := toMap(x.Args)
		if err != nil {
			return err
		}
		rhs, err := toMap(y.Args)
		if err != nil {
			return err
		}

		if err := diff.ANSI(lhs, rhs); err != nil {
			label := "Args"
			if opt.Name != "" {
				label = fmt.Sprintf("%s %s", label, opt.Name)
			}
			return fmt.Errorf("%s: %w", label, err)
		}
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

// Compare checks if two queries are equal.
func Compare(a, b string) (bool, error) {
	parser := sqlparser.NewTestParser()
	// To ignore variables, we can redact the statement.
	// parser.RedactSQLQuery(a)
	return parser.QueryMatchesTemplates(a, []string{b})
}

func normalize(q string) (string, error) {
	parser := sqlparser.NewTestParser()
	stmt, err := parser.Parse(q)
	if err != nil {
		return "", err
	}

	q = sqlparser.String(stmt)

	// sqlparser replaces all ? with the format :v1, :v2,
	// :vn ...
	return q, nil
}
