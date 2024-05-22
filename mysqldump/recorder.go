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
func NewRecorder(t *testing.T, db dbtx) *Recorder {
	d := &Recorder{
		t:    t,
		db:   db,
		opts: make(map[int][]Option),
		seen: make(map[string]int),
	}
	t.Cleanup(d.dump)
	return d
}

var _ dbtx = (*Recorder)(nil)

// Recorder logs the query and args.
type Recorder struct {
	t     *testing.T
	id    int
	db    dbtx
	opts  map[int][]Option
	seen  map[string]int
	dumps []*SQL
}

// SetDB sets the db.
func (r *Recorder) SetDB(db dbtx) {
	r.db = db
}

// SetCallOptions sets the options for the id-th call.
func (r *Recorder) SetOption(id int, opts ...Option) {
	r.opts[id] = opts
}

// Options returns the options for the id-th call.
func (r *Recorder) Options(id int) []Option {
	return r.opts[id]
}

func (r *Recorder) log(method, query string, args ...any) {
	defer func() {
		r.id++
	}()

	fileName := method
	r.seen[fileName]++
	if n := r.seen[fileName]; n > 1 {
		fileName = fmt.Sprintf("%s#%d", fileName, n)
	} else {
		fileName = fmt.Sprintf("%s#%d", fileName, 1)
	}

	r.opts[r.id] = append(r.opts[r.id], File(fileName))
	r.dumps = append(r.dumps, &SQL{
		Args:  args,
		Query: query,
	})
}

func (r *Recorder) dump() {
	for i, dump := range r.dumps {
		Dump(r.t, dump, r.Options(i)...)
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
