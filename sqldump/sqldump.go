package sqldump

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
)

func Query(ctx context.Context, db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	// Query all rows.
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get the column names.
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Ensure unique column names.
	columns = uniqueColumns(columns)

	// Serialize the rows into a slice of maps.
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
				if json.Valid(b) {
					var a any
					if err := json.Unmarshal(b, &a); err != nil {
						return nil, err
					}
					m[columns[i]] = a
				} else {
					m[columns[i]] = string(b)
				}
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
