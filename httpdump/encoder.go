package httpdump

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/alextanhongpin/testdump/httpdump/internal"
	"golang.org/x/tools/txtar"
)

type encoder struct {
	marshalFns []Transformer
	indentJSON bool
}

func (e *encoder) Marshal(v any) ([]byte, error) {
	h := v.(*HTTP)
	hc, err := h.Clone()
	if err != nil {
		return nil, err
	}

	for _, fn := range e.marshalFns {
		if err := fn(hc.Response, hc.Request); err != nil {
			return nil, err
		}
	}

	return Write(hc, e.indentJSON)
}

func (e *encoder) Unmarshal(b []byte) (any, error) {
	return Read(b)
}

// Write writes the request/response pair to bytes.
func Write(h *HTTP, pretty bool) ([]byte, error) {
	// Format the JSON body.
	internal.NormalizeRequest(h.Request)
	req, err := httputil.DumpRequest(h.Request, false)
	if err != nil {
		return nil, err
	}
	rc, err := internal.CloneRequest(h.Request)
	if err != nil {
		return nil, err
	}
	reqBody, err := io.ReadAll(rc.Body)
	if err != nil {
		return nil, err
	}

	res, err := httputil.DumpResponse(h.Response, false)
	if err != nil {
		return nil, err
	}
	wc, err := internal.CloneResponse(h.Response)
	if err != nil {
		return nil, err
	}
	resBody, err := io.ReadAll(wc.Body)
	if err != nil {
		return nil, err
	}

	if pretty {
		reqBody = internal.MustPrettyBytes(reqBody)
		resBody = internal.MustPrettyBytes(resBody)
	}
	if len(reqBody) == 0 {
		reqBody = append(reqBody, '\r', '\n')
	} else {
		reqBody = append(reqBody, '\r', '\n', '\r', '\n')
	}

	return txtar.Format(
		&txtar.Archive{
			Files: []txtar.File{
				{
					Name: requestFile,
					Data: req,
				},
				{
					Name: requestBodyFile,
					Data: reqBody,
				},
				{
					Name: responseFile,
					Data: res,
				},
				{
					Name: responseBodyFile,
					Data: resBody,
				},
			},
		},
	), nil
}

// Read reads the request/response pair from bytes.
func Read(b []byte) (*HTTP, error) {
	archive := txtar.Parse(b)

	h := new(HTTP)
	var err error
	for _, f := range archive.Files {
		switch f.Name {
		case requestFile:
			b := bufio.NewReader(bytes.NewReader(f.Data))
			h.Request, err = http.ReadRequest(b)
			if err != nil {
				return nil, err
			}
		case requestBodyFile:
			h.RequestBody = f.Data
		case responseFile:
			b := bufio.NewReader(bytes.NewReader(f.Data))
			h.Response, err = http.ReadResponse(b, nil)
			if err != nil {
				return nil, err
			}
		case responseBodyFile:
			h.ResponseBody = f.Data
		}
	}

	return h, nil
}
