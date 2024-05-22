package pgdump_test

import (
	"testing"

	"github.com/alextanhongpin/dump/pgdump"
)

func TestDump(t *testing.T) {
	dump := &pgdump.SQL{
		Query: `select * from users where name = $1 and age = $2`,
		Args:  []any{"John", 13},
	}

	pgdump.Dump(t, dump)
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
