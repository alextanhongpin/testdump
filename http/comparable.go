package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alextanhongpin/dump/http/internal"
	"github.com/alextanhongpin/dump/pkg/diff"
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

func (c CompareOption) Merge(o CompareOption) CompareOption {
	return CompareOption{
		Header:  append(c.Header, o.Header...),
		Body:    append(c.Body, o.Body...),
		Trailer: append(c.Trailer, o.Trailer...),
	}
}

func (x *Comparable) Compare(y *Comparable, opt CompareOption) error {
	if err := diff.ANSI(x.Line, y.Line); err != nil {
		return fmt.Errorf("Line: %w", err)
	}

	if err := diff.ANSI(x.Body, y.Body, opt.Body...); err != nil {
		return fmt.Errorf("Body: %w", err)
	}

	if err := diff.ANSI(x.Header, y.Header, opt.Header...); err != nil {
		return fmt.Errorf("Header: %w", err)
	}

	if err := diff.ANSI(x.Trailer, y.Trailer, opt.Trailer...); err != nil {
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
		b = bytes.TrimSpace(b)

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
	b = bytes.TrimSpace(b)

	var a any
	if json.Valid(b) {
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
	} else {
		a = string(b)
	}

	r.Body = io.NopCloser(bytes.NewReader(b))

	return &Comparable{
		Line:    internal.FormatResponseLine(r),
		Header:  r.Header.Clone(),
		Body:    a,
		Trailer: r.Trailer.Clone(),
	}, nil
}

func CompareRequest(s, t *http.Request, opt CompareOption) error {
	lhs, err := NewComparableRequest(s)
	if err != nil {
		return err
	}

	rhs, err := NewComparableRequest(t)
	if err != nil {
		return err
	}

	return lhs.Compare(rhs, opt)
}

func CompareResponse(s, t *http.Response, opt CompareOption) error {
	lhs, err := NewComparableResponse(s)
	if err != nil {
		return err
	}

	rhs, err := NewComparableResponse(t)
	if err != nil {
		return err
	}

	return lhs.Compare(rhs, opt)
}
