package mysqldump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/tools/txtar"
	"vitess.io/vitess/go/vt/sqlparser"
)

const (
	querySection = "query"
	argsSection  = "args"
)

type encoder struct {
	marshalFns []func(*SQL) error
}

func (e *encoder) Marshal(v any) ([]byte, error) {
	return Write(v.(*SQL), e.marshalFns...)
}

func (e *encoder) Unmarshal(b []byte) (a any, err error) {
	return Read(b)
}

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
					key := strings.ReplaceAll(k, ":v", "")
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

	// Convert args into key-value pairs.
	// Each key is :v1, :v2, :vn ...
	var a []byte
	if len(sql.Args) > 0 {
		args := make(map[string]any)
		for i, v := range sql.Args {
			// sqlparser replaces all '?' with ':v1', ':v2', ':vn'
			// ...
			k := fmt.Sprintf(":v%d", i+1)
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

func normalize(q string) (string, error) {
	parser := sqlparser.NewTestParser()
	stmt, err := parser.Parse(q)
	if err != nil {
		return "", err
	}

	q = sqlparser.String(stmt)

	// sqlparser replaces all ? with the format :v1, :v2,
	// :vn ...
	return q, nil
}
