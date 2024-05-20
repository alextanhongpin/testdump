package mysqldump_test

import (
	"testing"

	"github.com/alextanhongpin/dump/mysqldump"
)

func TestDump(t *testing.T) {
	dump := &mysqldump.SQL{
		Query: `select * from users where name = ? and age = ?`,
		Args:  []any{"John", 13},
	}

	mysqldump.Dump(t, dump)
}

func TestCompare(t *testing.T) {
	base := `select name from users where name = ?`
	vars := []string{
		base,
		"SELECT `name` FROM `users` WHERE name = ?",
		`select name
from users
		where name = ?`,
	}

	for _, v := range vars {
		ok, err := mysqldump.Compare(base, v)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Errorf("expected %s to match %s", base, v)
		}
	}
}
