package mysqldump_test

import (
	"testing"
	"time"

	"github.com/alextanhongpin/testdump/mysqldump"
)

func TestDump(t *testing.T) {
	dump := &mysqldump.SQL{
		Query: `select * from users where name = ? and age = ?`,
		Args:  []any{"John", 13},
	}

	mysqldump.Dump(t, dump)
}

func TestIgnoreFields(t *testing.T) {
	dump := &mysqldump.SQL{
		Query: `select * from users where name = ? and created_at > ?`,
		Args:  []any{"John", time.Now()},
	}

	mysqldump.Dump(t, dump, mysqldump.IgnoreArgs(":v2"))
}

func TestTransformer(t *testing.T) {
	dump := &mysqldump.SQL{
		Query: `select * from users where name = ? and id = ?`,
		Args:  []any{"John", 1},
	}

	// Add pretty print sql for all dumps.
	md := mysqldump.New(mysqldump.Prettify)
	md.Dump(t, dump)
}

func TestCompareQuery(t *testing.T) {
	base := `select name from users where name = ?`
	vars := []string{
		base,
		"SELECT `name` FROM `users` WHERE name = ?",
		`select name
from users
		where name = ?`,
	}

	for _, v := range vars {
		ok, err := mysqldump.CompareQuery(base, v)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Errorf("expected %s to match %s", base, v)
		}
	}
}
