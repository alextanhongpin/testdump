package pgdump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v4"
	"golang.org/x/tools/txtar"
)

const (
	querySection = "query"
	argsSection  = "args"
)

func Read(b []byte) (*SQL, error) {
	d := new(SQL)

	arc := txtar.Parse(b)
	for _, f := range arc.Files {
		name, data := f.Name, bytes.TrimSpace(f.Data)

		switch name {
		case querySection:
			d.Query = string(data)
		case argsSection:
			var a any
			if err := json.Unmarshal(data, &a); err != nil {
				return nil, err
			}

			m, ok := a.(map[string]any)
			if ok {
				// Sort to ensure the order is consistent.
				d.Args = make([]any, len(m))
				for k, v := range m {
					key := strings.ReplaceAll(k, "$", "")
					i, err := strconv.Atoi(key) // Index starts at 1
					if err != nil {
						return nil, err
					}
					d.Args[i-1] = v
				}
			}
		}
	}

	return d, nil
}

func Write(sql *SQL, transformers ...func(*SQL) error) ([]byte, error) {
	q, err := normalize(sql.Query)
	if err != nil {
		return nil, err
	}
	sql.Query = q

	for _, transform := range transformers {
		if err := transform(sql); err != nil {
			return nil, err
		}
	}

	var a []byte
	if len(sql.Args) > 0 {
		args := make(map[string]any)
		for i, v := range sql.Args {
			k := fmt.Sprintf("$%d", i+1)
			args[k] = v
		}

		a, err = json.MarshalIndent(args, "", " ")
		if err != nil {
			return nil, err
		}
	}

	arc := new(txtar.Archive)
	// Query.
	arc.Files = append(arc.Files, txtar.File{
		Name: querySection,
		Data: appendNewLine([]byte(sql.Query)),
	})

	// Args.
	if len(a) != 0 {
		arc.Files = append(arc.Files, txtar.File{
			Name: argsSection,
			Data: appendNewLine(a),
		})
	}

	return txtar.Format(arc), nil
}

func appendNewLine(b []byte) []byte {
	b = append(b, '\n')
	b = append(b, '\n')
	return b
}

// normalize standardize the capitalization and strip of new lines etc.
func normalize(q string) (string, error) {
	stmt, err := pg_query.Parse(q)
	if err != nil {
		return "", err
	}

	q, err = pg_query.Deparse(stmt)
	if err != nil {
		return "", err
	}

	return q, nil
}
