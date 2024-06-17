package sqldump_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alextanhongpin/testdump/jsondump"
	"github.com/alextanhongpin/testdump/sqldump"
)

func TestDump(t *testing.T) {
	db := newMockDB(t,
		[]string{"id", "name", "email"},
		[]string{
			"1", "alpha", "alpha@mail.com",
			"1", "beta", "beta@mail.com",
			"1", "charlie", "charlie@mail.com",
			"1", "dwayne", "dwayne@mail.com",
			"1", "ellie", "ellie@mail.com",
		},
	)

	res, err := sqldump.Dump(context.Background(), db, `select * from users`)
	if err != nil {
		t.Fatal(err)
	}
	jsondump.Dump(t, res)
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
