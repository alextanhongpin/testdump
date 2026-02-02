package httpdump

import (
	"bytes"
	"net/http"

	"github.com/alextanhongpin/testdump/httpdump/internal"
)

const (
	requestFile      = "request.http"
	requestBodyFile  = "request_body.http"
	responseFile     = "response.http"
	responseBodyFile = "response_body.http"
)

type HTTP struct {
	Request      *http.Request
	RequestBody  []byte
	Response     *http.Response
	ResponseBody []byte
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
		Request:      r,
		RequestBody:  bytes.Clone(h.RequestBody),
		Response:     w,
		ResponseBody: bytes.Clone(h.ResponseBody),
	}, nil
}

// Message is a comparable representation of the request/response pair.
type Message struct {
	Line    string      `json:"line"`
	Header  http.Header `json:"header"`
	Body    any         `json:"body"`
	Trailer http.Header `json:"trailer"`
}
