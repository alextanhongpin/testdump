package http

import (
	"net/http"

	"golang.org/x/tools/txtar"
)

const requestFile = "request.http"
const responseFile = "response.http"

type Dump struct {
	Line    string      `json:"line"`
	Header  http.Header `json:"header"`
	Body    any         `json:"body"`
	Trailer http.Header `json:"trailer"`
}

func Write(w *http.Response, r *http.Request) ([]byte, error) {
	req, err := WriteRequest(r)
	if err != nil {
		return nil, err
	}

	res, err := WriteResponse(w)
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

func Read(b []byte) (w *http.Response, r *http.Request, err error) {
	archive := txtar.Parse(b)

	for _, f := range archive.Files {
		switch f.Name {
		case requestFile:
			r, err = ReadRequest(f.Data)
			if err != nil {
				return nil, nil, err
			}
		case responseFile:
			w, err = ReadResponse(f.Data)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return
}
