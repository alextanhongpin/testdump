package pgdump

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

type dbtx interface {
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Recorder logs the query and args.
type Recorder struct {
	dumps    []*SQL
	id       int
	opts     []Option
	optsByID map[int][]Option
	seen     map[string]int
	t        *testing.T
}

// NewRecorder ...
func NewRecorder(t *testing.T, opts ...Option) *Recorder {
	d := &Recorder{
		t:        t,
		opts:     opts,
		optsByID: make(map[int][]Option),
		seen:     make(map[string]int),
	}
	t.Cleanup(d.dump)
	return d
}

// SetOptionsAt sets the options for the id-th call.
func (r *Recorder) SetOptionsAt(id int, opts ...Option) {
	r.optsByID[id] = opts
}

func (r *Recorder) Record(method, query string, args ...any) {
	fileName := method
	r.seen[fileName]++
	fileName = fmt.Sprintf("%s#%d", fileName, r.seen[fileName])

	r.optsByID[r.id] = append(r.optsByID[r.id], File(fileName))
	r.dumps = append(r.dumps, &SQL{
		Args:  args,
		Query: query,
	})
	r.id++
}

func (r *Recorder) DB(db dbtx) dbtx {
	return NewDBRecorder(db, r)
}

func (r *Recorder) dump() {
	for i, dump := range r.dumps {
		Dump(r.t, dump, append(r.opts, r.optsByID[i]...)...)
	}
}

type recorder interface {
	Record(method, query string, args ...any)
}

var _ dbtx = (*DB)(nil)

type DB struct {
	rec recorder
	db  dbtx
}

func NewDBRecorder(db dbtx, rec recorder) *DB {
	return &DB{db: db, rec: rec}
}

func (d *DB) SetDB(db dbtx) {
	d.db = db
}

func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	d.rec.Record("exec", query, args...)

	return d.db.Exec(query, args...)
}

func (d *DB) Prepare(query string) (*sql.Stmt, error) {
	d.rec.Record("prepare", query)

	return d.db.Prepare(query)
}

func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	d.rec.Record("query", query, args...)

	return d.db.Query(query, args...)
}

func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	d.rec.Record("query_row", query, args...)

	return d.db.QueryRow(query, args...)
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	d.rec.Record("exec_context", query, args...)

	return d.db.ExecContext(ctx, query, args...)
}

func (d *DB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	d.rec.Record("prepare_context", query)

	return d.db.PrepareContext(ctx, query)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	d.rec.Record("query_context", query, args...)

	return d.db.QueryContext(ctx, query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	d.rec.Record("query_row_context", query, args...)

	return d.db.QueryRowContext(ctx, query, args...)
}
