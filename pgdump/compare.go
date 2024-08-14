package pgdump

import (
	"encoding/json"
	"fmt"

	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/google/go-cmp/cmp"
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

type comparer struct {
	opts   []cmp.Option
	colors bool
	file   string
}

func (c *comparer) Compare(a, b any) error {
	x := a.(*SQL)
	y := b.(*SQL)

	err := c.compare(x, y)
	if err != nil {
		if c.file != "" {
			return fmt.Errorf("%s: %w", c.file, err)
		}

		return err
	}

	return nil
}

func (c *comparer) compare(snapshot, received *SQL) error {
	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

	ok, err := CompareQuery(snapshot.Query, received.Query)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("Query: %w", comparer(snapshot.Query, received.Query))
	}

	lhs, err := toMap(snapshot.Args)
	if err != nil {
		return err
	}
	rhs, err := toMap(received.Args)
	if err != nil {
		return err
	}

	if err := comparer(lhs, rhs, c.opts...); err != nil {
		return fmt.Errorf("Args: %w", err)
	}

	return nil
}

type SQL struct {
	Query string
	Args  []any
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
