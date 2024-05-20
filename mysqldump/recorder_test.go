package mysqldump_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alextanhongpin/dump/mysqldump"
)

func TestRecorder(t *testing.T) {
	db := newMockDB(t,
		[]string{"id", "name"}, // Cols
		"1", "Alice",           // Two mock rows
		"2", "Bob",
	)

	rec := mysqldump.NewRecorder(t, db)
	ctx := context.Background()

	var id int
	var name string
	if err := rec.QueryRowContext(ctx, "select * from users where id = ?", 1).Scan(&id, &name); err != nil {
		t.Fatal(err)
	}
	if id != 1 || name != "Alice" {
		t.Errorf("expected 1, Alice, got %d, %s", id, name)
	}

	if err := rec.QueryRowContext(ctx, "select * from users where id = ?", 2).Scan(&id, &name); err != nil {
		t.Fatal(err)
	}
	if id != 2 || name != "Bob" {
		t.Errorf("expected 2, Bob, got %d, %s", id, name)
	}
}

func newMockDB(t *testing.T, cols []string, vals ...string) *sql.DB {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if len(vals)%len(cols) != 0 {
		panic("invalid row values")
	}

	// Allow execution of multiple queries.
	for i := 0; i < len(vals); i += len(cols) {
		dvals := make([]driver.Value, len(cols))
		rowvals := vals[i : i+len(cols)]
		for i := range len(cols) {
			dvals[i] = rowvals[i]
		}

		rows := sqlmock.NewRows(cols).AddRow(dvals...)
		mock.ExpectQuery("select(.+)").WillReturnRows(rows)
	}

	return db
}
