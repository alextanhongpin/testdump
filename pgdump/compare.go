package pgdump

import (
	"fmt"

	"github.com/alextanhongpin/dump/pkg/diff"
	pg_query "github.com/pganalyze/pg_query_go/v4"
)

func Compare(snapshot, received *SQL) error {
	ok, err := CompareQuery(snapshot.Query, received.Query)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("Query: %w", diff.ANSI(snapshot.Query, received.Query))
	}

	lhs, err := toMap(snapshot.Args)
	if err != nil {
		return err
	}
	rhs, err := toMap(received.Args)
	if err != nil {
		return err
	}

	if err := diff.ANSI(lhs, rhs); err != nil {
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
