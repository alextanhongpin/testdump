package httpdump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alextanhongpin/testdump/httpdump/internal"
	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/google/go-cmp/cmp"
)

type comparer struct {
	colors bool
	cmpOpt CompareOption
}

func (c *comparer) Compare(a, b any) error {
	x := a.(*HTTP)
	y := b.(*HTTP)

	return c.compare(x, y)
}

func (c *comparer) compare(snapshot, received *HTTP) error {
	if err := c.compareRequest(snapshot.Request, received.Request); err != nil {
		return fmt.Errorf("Request: %w", err)
	}

	if err := c.compareResponse(snapshot.Response, received.Response); err != nil {
		return fmt.Errorf("Response: %w", err)
	}

	return nil
}

func (c *comparer) compareRequest(snapshot, received *http.Request) error {
	s, err := NewComparableRequest(snapshot)
	if err != nil {
		return err
	}

	r, err := NewComparableRequest(received)
	if err != nil {
		return err
	}

	return c.compareMessage(s, r, c.cmpOpt.Request)
}

func (c *comparer) compareResponse(snapshot, received *http.Response) error {
	s, err := NewComparableResponse(snapshot)
	if err != nil {
		return err
	}

	r, err := NewComparableResponse(received)
	if err != nil {
		return err
	}

	return c.compareMessage(s, r, c.cmpOpt.Response)
}

func (c *comparer) compareMessage(snapshot, received *Message, opt CompareMessageOption) error {
	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

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
