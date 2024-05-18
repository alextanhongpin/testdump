package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alextanhongpin/dump/http/internal"
	"github.com/google/go-cmp/cmp"
)

const (
	requestFile  = "request.http"
	responseFile = "response.http"
)

// Comparable is a comparable representation of the request/response pair.
type Comparable struct {
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

func (x *Comparable) Compare(y *Comparable, opt *CompareOption) error {
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

func NewComparableRequest(r *http.Request) (*Comparable, error) {
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

	return &Comparable{
		Line:   internal.FormatRequestLine(r),
		Header: r.Header.Clone(),
		Body:   a,
	}, nil
}

func NewComparableResponse(r *http.Response) (*Comparable, error) {
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

	return &Comparable{
		Line:    internal.FormatResponseLine(r),
		Header:  r.Header.Clone(),
		Body:    a,
		Trailer: r.Trailer.Clone(),
	}, nil
}
