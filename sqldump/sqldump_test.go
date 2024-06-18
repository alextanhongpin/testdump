package sqldump_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alextanhongpin/testdump/sqldump"
	"github.com/alextanhongpin/testdump/yamldump"
)

func TestDump(t *testing.T) {
	db := newMockDB(t,
		[]string{"id", "name", "email", "created_at"},
		[]string{
			"1", "alpha", "alpha@mail.com", time.Now().String(),
			"1", "beta", "beta@mail.com", time.Now().String(),
			"1", "charlie", "charlie@mail.com", time.Now().String(),
			"1", "dwayne", "dwayne@mail.com", time.Now().String(),
			"1", "ellie", "ellie@mail.com", time.Now().String(),
		},
	)

	res, err := sqldump.Dump(context.Background(), db, `select * from users`)
	if err != nil {
		t.Fatal(err)
	}
	yamldump.Dump(t, res, yamldump.IgnoreFields("created_at"))
}

func newMockDB(t *testing.T, cols []string, vals []string) *sql.DB {
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

	dvals := make([]driver.Value, len(vals))
	for i := range vals {
		dvals[i] = vals[i]
	}

	rows := sqlmock.NewRows(cols)
	for i := 0; i < len(vals); i += len(cols) {
		rows.AddRow(dvals[i : i+len(cols)]...)
	}

	mock.ExpectQuery("select(.+)").WillReturnRows(rows)

	return db
}
