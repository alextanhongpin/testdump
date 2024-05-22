package pgdump_test

import (
	"testing"
	"time"

	"github.com/alextanhongpin/dump/pgdump"
	"github.com/alextanhongpin/dump/pkg/sqlformat"
)

func TestDump(t *testing.T) {
	dump := &pgdump.SQL{
		Query: `select * from users where name = $1 and age = $2`,
		Args:  []any{"John", 13},
	}

	pgdump.Dump(t, dump)
}

func TestIgnoreFields(t *testing.T) {
	dump := &pgdump.SQL{
		Query: `select * from users where name = $1 and created_at > $2`,
		Args:  []any{"John", time.Now()},
	}

	pgdump.Dump(t, dump, pgdump.IgnoreArgs("$2"))
}

func TestTransformer(t *testing.T) {
	prettySQL := func(s *pgdump.SQL) error {
		q, err := sqlformat.Format(s.Query)
		if err != nil {
			return err
		}
		s.Query = q
		return nil
	}

	dump := &pgdump.SQL{
		Query: `select * from users where name = $1 and id = $2`,
		Args:  []any{"John", 1},
	}

	// Add pretty print sql for all dumps.
	pd := pgdump.New(pgdump.Transformers(prettySQL))
	pd.Dump(t, dump)
}

func TestCompareQuery(t *testing.T) {
	base := `select name from users where name = $1`
	vars := []string{
		base,
		`SELECT "name" FROM "users" WHERE name = $1`,
		`select name
from users
		where name = $1`,
	}

	for _, v := range vars {
		ok, err := pgdump.CompareQuery(base, v)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Errorf("expected %s to match %s", base, v)
		}
	}
}
