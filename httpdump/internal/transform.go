package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func MustPrettyBytes(b []byte) []byte {
	b, err := PrettyBytes(b)
	if err != nil {
		panic(fmt.Errorf("failed to indent json: %w", err))
	}
	return b
}

func PrettyBytes(b []byte) ([]byte, error) {
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
