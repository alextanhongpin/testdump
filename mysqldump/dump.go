package mysqldump

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/alextanhongpin/dump/pkg/sqlformat"
	"golang.org/x/tools/txtar"
)

const (
	querySection = "query"
	argsSection  = "args"
)

type SQL struct {
	Query string
	Args  []any
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
				for _, v := range m {
					d.Args = append(d.Args, v)
				}
			}
		}
	}

	return d, nil
}

func Write(sql *SQL) ([]byte, error) {
	q, err := normalize(sql.Query)
	if err != nil {
		return nil, err
	}

	q, err = sqlformat.Format(q)
	if err != nil {
		return nil, err
	}

	// Convert args into key-value pairs.
	// Each key is :v1, :v2, :vn ...
	var a []byte
	if len(sql.Args) > 0 {
		args := make(map[string]any)
		for i, v := range sql.Args {
			// sqlparser replaces all '?' with ':v1', ':v2', ':vn'
			// ...
			k := fmt.Sprintf("v%d", i+1)
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
		Data: appendNewLine([]byte(q)),
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
