package mysqldump

import (
	"fmt"

	"github.com/alextanhongpin/dump/pkg/diff"
	"vitess.io/vitess/go/vt/sqlparser"
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

// CompareQuery checks if two queries are equal.
func CompareQuery(a, b string) (bool, error) {
	parser := sqlparser.NewTestParser()
	// To ignore variables, we can redact the statement.
	// parser.RedactSQLQuery(a)
	return parser.QueryMatchesTemplates(a, []string{b})
}
