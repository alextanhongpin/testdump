package httpdump

import (
	"net/http"

	"github.com/alextanhongpin/testdump/httpdump/internal"
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

// Message is a comparable representation of the request/response pair.
type Message struct {
	Line    string      `json:"line"`
	Header  http.Header `json:"header"`
	Body    any         `json:"body"`
	Trailer http.Header `json:"trailer"`
}
