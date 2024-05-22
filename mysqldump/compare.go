package mysqldump

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"vitess.io/vitess/go/vt/sqlparser"
)

type SQL struct {
	Query string
	Args  []any
}

type CompareOption struct {
	CmpOpts []cmp.Option
}

type comparer func(a, b any, opts ...cmp.Option) error

func (snapshot *SQL) Compare(received *SQL, opt CompareOption, cmp comparer) error {
	ok, err := CompareQuery(snapshot.Query, received.Query)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("Query: %w", cmp(snapshot.Query, received.Query))
	}

	lhs, err := toMap(snapshot.Args)
	if err != nil {
		return err
	}
	rhs, err := toMap(received.Args)
	if err != nil {
		return err
	}

	if err := cmp(lhs, rhs, opt.CmpOpts...); err != nil {
		return fmt.Errorf("Args: %w", err)
	}

	return nil
}

// CompareQuery checks if two queries are equal.
func CompareQuery(a, b string) (bool, error) {
	parser := sqlparser.NewTestParser()
	// To ignore variables, we can redact the statement.
	// parser.RedactSQLQuery(a)
	return parser.QueryMatchesTemplates(a, []string{b})
}

// toMap converts the slice args into a map for better diff.
// Each key is named `:vn`, where `n` indicates the index of the arg in the
// slice.
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
