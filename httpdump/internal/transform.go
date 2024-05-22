package internal

import (
	"bytes"
	"encoding/json"
)

func prettyBytes(b []byte) ([]byte, error) {
	if !json.Valid(b) {
		return b, nil
	}

	bb := new(bytes.Buffer)
	if err := json.Indent(bb, b, "", " "); err != nil {
		return nil, err
	}

	b = bb.Bytes()

	return b, nil
}
