package mysqldump

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

// NewRecorder ...
func NewRecorder(t *testing.T, db dbtx, opts ...Option) *Recorder {
	d := &Recorder{
		t:        t,
		db:       db,
		opts:     opts,
		optsByID: make(map[int][]Option),
		seen:     make(map[string]int),
	}
	t.Cleanup(d.dump)
	return d
}

var _ dbtx = (*Recorder)(nil)

// Recorder logs the query and args.
type Recorder struct {
	db       dbtx
	dumps    []*SQL
	id       int
	opts     []Option
	optsByID map[int][]Option
	seen     map[string]int
	t        *testing.T
}

// SetDB sets the db.
func (r *Recorder) SetDB(db dbtx) {
	r.db = db
}

// SetOptionsAt sets the options for the id-th call.
func (r *Recorder) SetOptionsAt(id int, opts ...Option) {
	r.optsByID[id] = opts
}

func (r *Recorder) log(method, query string, args ...any) {
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

func (r *Recorder) dump() {
	for i, dump := range r.dumps {
		Dump(r.t, dump, append(r.opts, r.optsByID[i]...)...)
	}
}

func (r *Recorder) Exec(query string, args ...any) (sql.Result, error) {
	r.log("exec", query, args...)

	return r.db.Exec(query, args...)
}

func (r *Recorder) Prepare(query string) (*sql.Stmt, error) {
	r.log("prepare", query)

	return r.db.Prepare(query)
}

func (r *Recorder) Query(query string, args ...any) (*sql.Rows, error) {
	r.log("query", query, args...)

	return r.db.Query(query, args...)
}

func (r *Recorder) QueryRow(query string, args ...any) *sql.Row {
	r.log("query_row", query, args...)

	return r.db.QueryRow(query, args...)
}

func (r *Recorder) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	r.log("exec_context", query, args...)

	return r.db.ExecContext(ctx, query, args...)
}

func (r *Recorder) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	r.log("prepare_context", query)

	return r.db.PrepareContext(ctx, query)
}

func (r *Recorder) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	r.log("query_context", query, args...)

	return r.db.QueryContext(ctx, query, args...)
}

func (r *Recorder) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	r.log("query_row_context", query, args...)

	return r.db.QueryRowContext(ctx, query, args...)
}
