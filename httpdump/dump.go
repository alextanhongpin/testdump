package httpdump

import (
	"bufio"
	"bytes"
	"net/http"

	"github.com/alextanhongpin/dump/httpdump/internal"
	"golang.org/x/tools/txtar"
)

// Write writes the request/response pair to bytes.
func Write(w *http.Response, r *http.Request, pretty bool) ([]byte, error) {
	// Format the JSON body.
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
