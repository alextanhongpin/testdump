package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alextanhongpin/dump/http/internal"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

const requestFile = "request.http"
const responseFile = "response.http"

// Dump is a comparable representation of the request/response pair.
type Dump struct {
	Line    string      `json:"line"`
	Header  http.Header `json:"header"`
	Body    any         `json:"body"`
	Trailer http.Header `json:"trailer"`
}

type CompareOption struct {
	Header  []cmp.Option
	Body    []cmp.Option
	Trailer []cmp.Option
}

func (x *Dump) Compare(y *Dump, opt *CompareOption) error {
	if opt == nil {
		opt = new(CompareOption)
	}

	if err := ANSIDiff(x.Line, y.Line); err != nil {
		return fmt.Errorf("Line: %w", err)
	}

	if err := ANSIDiff(x.Body, y.Body, opt.Body...); err != nil {
		return fmt.Errorf("Body: %w", err)
	}

	if err := ANSIDiff(x.Header, y.Header, opt.Header...); err != nil {
		return fmt.Errorf("Header: %w", err)
	}

	if err := ANSIDiff(x.Trailer, y.Trailer, opt.Trailer...); err != nil {
		return fmt.Errorf("Trailer: %w", err)
	}

	return nil
}

func NewRequestDump(r *http.Request) (*Dump, error) {
	var a any
	if r.Body != nil {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		if json.Valid(b) {
			if err := json.Unmarshal(b, &a); err != nil {
				return nil, err
			}
		} else {
			a = string(b)
		}

		r.Body = io.NopCloser(bytes.NewReader(b))
	}

	return &Dump{
		Line:   internal.FormatRequestLine(r),
		Header: r.Header.Clone(),
		Body:   a,
	}, nil
}

func NewResponseDump(r *http.Response) (*Dump, error) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var a any
	if json.Valid(b) {
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
	} else {
		a = string(bytes.TrimSpace(b))
	}

	b = bytes.TrimSpace(b)
	r.Body = io.NopCloser(bytes.NewReader(b))

	return &Dump{
		Line:    internal.FormatResponseLine(r),
		Header:  r.Header.Clone(),
		Body:    a,
		Trailer: r.Trailer.Clone(),
	}, nil
}

// Write writes the request/response pair to bytes.
func Write(w *http.Response, r *http.Request) ([]byte, error) {
	// Format the JSON body.
	pretty := true
	req, err := internal.DumpRequest(r, pretty)
	if err != nil {
		return nil, err
	}

	res, err := internal.DumpResponse(w, pretty)
	if err != nil {
		return nil, err
	}

	return txtar.Format(
		&txtar.Archive{
			Files: []txtar.File{
				{
					Name: requestFile,
					Data: req,
				},
				{
					Name: responseFile,
					Data: res,
				},
			},
		},
	), nil
}

// Read reads the request/response pair from bytes.
func Read(b []byte) (w *http.Response, r *http.Request, err error) {
	archive := txtar.Parse(b)

	for _, f := range archive.Files {
		b := bufio.NewReader(bytes.NewReader(f.Data))
		switch f.Name {
		case requestFile:
			r, err = http.ReadRequest(b)
			if err != nil {
				return nil, nil, err
			}
		case responseFile:
			w, err = http.ReadResponse(b, nil)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return
}
