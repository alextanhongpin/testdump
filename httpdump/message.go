package httpdump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alextanhongpin/dump/httpdump/internal"
	"github.com/google/go-cmp/cmp"
)

const (
	requestFile  = "request.http"
	responseFile = "response.http"
)

type HTTP struct {
	Request  *http.Request
	Response *http.Response
}

func (h *HTTP) Clone() (*HTTP, error) {
	r, err := internal.CloneRequest(h.Request)
	if err != nil {
		return nil, err
	}

	w, err := internal.CloneResponse(h.Response)
	if err != nil {
		return nil, err
	}

	return &HTTP{
		Request:  r,
		Response: w,
	}, nil
}

func (snapshot *HTTP) Compare(received *HTTP, opt CompareOption, comparer func(a, b any, opts ...cmp.Option) error) error {
	if err := CompareRequest(snapshot.Request, received.Request, opt.Request, comparer); err != nil {
		return fmt.Errorf("Request: %w", err)
	}

	if err := CompareResponse(snapshot.Response, received.Response, opt.Response, comparer); err != nil {
		return fmt.Errorf("Response: %w", err)
	}

	return nil
}

type CompareOption struct {
	Request  CompareMessageOption
	Response CompareMessageOption
}

func (c CompareOption) Merge(o CompareOption) CompareOption {
	return CompareOption{
		Request:  c.Request.Merge(o.Request),
		Response: c.Response.Merge(o.Response),
	}
}

// Message is a comparable representation of the request/response pair.
type Message struct {
	Line    string      `json:"line"`
	Header  http.Header `json:"header"`
	Body    any         `json:"body"`
	Trailer http.Header `json:"trailer"`
}

type CompareMessageOption struct {
	Header  []cmp.Option
	Body    []cmp.Option
	Trailer []cmp.Option
}

func (c CompareMessageOption) Merge(o CompareMessageOption) CompareMessageOption {
	return CompareMessageOption{
		Header:  append(c.Header, o.Header...),
		Body:    append(c.Body, o.Body...),
		Trailer: append(c.Trailer, o.Trailer...),
	}
}

func (snapshot *Message) Compare(received *Message, opt CompareMessageOption, comparer func(a, b any, opts ...cmp.Option) error) error {
	if err := comparer(snapshot.Line, received.Line); err != nil {
		return fmt.Errorf("Line: %w", err)
	}

	if err := comparer(snapshot.Body, received.Body, opt.Body...); err != nil {
		return fmt.Errorf("Body: %w", err)
	}

	if err := comparer(snapshot.Header, received.Header, opt.Header...); err != nil {
		return fmt.Errorf("Header: %w", err)
	}

	if err := comparer(snapshot.Trailer, received.Trailer, opt.Trailer...); err != nil {
		return fmt.Errorf("Trailer: %w", err)
	}

	return nil
}

func NewComparableRequest(r *http.Request) (*Message, error) {
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

	return &Message{
		Line:   internal.FormatRequestLine(r),
		Header: r.Header.Clone(),
		Body:   a,
	}, nil
}

func NewComparableResponse(r *http.Response) (*Message, error) {
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

	return &Message{
		Line:    internal.FormatResponseLine(r),
		Header:  r.Header.Clone(),
		Body:    a,
		Trailer: r.Trailer.Clone(),
	}, nil
}

func CompareRequest(snapshot, received *http.Request, opt CompareMessageOption, comparer func(a, b any, opts ...cmp.Option) error) error {
	s, err := NewComparableRequest(snapshot)
	if err != nil {
		return err
	}

	r, err := NewComparableRequest(received)
	if err != nil {
		return err
	}

	return s.Compare(r, opt, comparer)
}

func CompareResponse(snapshot, received *http.Response, opt CompareMessageOption, comparer func(a, b any, opts ...cmp.Option) error) error {
	s, err := NewComparableResponse(snapshot)
	if err != nil {
		return err
	}

	r, err := NewComparableResponse(received)
	if err != nil {
		return err
	}

	return s.Compare(r, opt, comparer)
}
