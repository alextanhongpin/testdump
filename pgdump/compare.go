package pgdump

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

type SQL struct {
	Query string
	Args  []any
}

type CompareOption struct {
	CmpOpts []cmp.Option
}

func (snapshot *SQL) Compare(received *SQL, opt CompareOption, cmp func(a, b any, opts ...cmp.Option) error) error {
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

// CompareQuery checks if two queries are equal, ignoring variables.
func CompareQuery(a, b string) (bool, error) {
	fa, err := pg_query.Fingerprint(a)
	if err != nil {
		return false, err
	}
	fb, err := pg_query.Fingerprint(b)
	if err != nil {
		return false, err
	}

	return fa == fb, nil
}

// toMap converts the slice args into a map for better diff.
// Each key is named `$n`, where `n` indicates the index of the arg in the
// slice.
func toMap(s []any) (any, error) {
	m := make(map[string]any)
	for k, v := range s {
		m[fmt.Sprintf("$%d", k+1)] = v
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
