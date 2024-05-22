package httpdump

import (
	"bufio"
	"bytes"
	"net/http"

	"github.com/alextanhongpin/dump/httpdump/internal"
	"golang.org/x/tools/txtar"
)

// Write writes the request/response pair to bytes.
func Write(h *HTTP, pretty bool) ([]byte, error) {
	// Format the JSON body.
	req, err := internal.DumpRequest(h.Request, pretty)
	if err != nil {
		return nil, err
	}

	res, err := internal.DumpResponse(h.Response, pretty)
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
func Read(b []byte) (*HTTP, error) {
	archive := txtar.Parse(b)

	h := new(HTTP)
	var err error
	for _, f := range archive.Files {
		b := bufio.NewReader(bytes.NewReader(f.Data))
		switch f.Name {
		case requestFile:
			h.Request, err = http.ReadRequest(b)
			if err != nil {
				return nil, err
			}
		case responseFile:
			h.Response, err = http.ReadResponse(b, nil)
			if err != nil {
				return nil, err
			}
		}
	}

	return h, nil
}
