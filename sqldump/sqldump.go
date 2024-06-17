package sqldump

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
)

func Dump(ctx context.Context, db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	columns = uniqueColumns(columns)

	var result []map[string]any
	for rows.Next() {
		cols := make([]any, len(columns))
		for i := range cols {
			cols[i] = &cols[i]
		}

		if err := rows.Scan(cols...); err != nil {
			return nil, err
		}
		m := make(map[string]any)
		for i := range columns {
			if b, ok := cols[i].([]byte); ok {
				m[columns[i]] = string(b)
			} else {
				m[columns[i]] = cols[i]
			}
		}
		result = append(result, m)
	}

	return result, nil
}

func uniqueColumns(columns []string) []string {
	columns = slices.Clone(columns)
	m := make(map[string]int)
	for i, c := range columns {
		v, ok := m[c]
		if ok {
			columns[i] = fmt.Sprintf("%s%d", c, v)
			m[c]++
		} else {
			m[c] = 1
		}
	}
	return columns
}
