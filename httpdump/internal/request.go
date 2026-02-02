package internal

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"net/http"
)

func CloneRequest(r *http.Request) (*http.Request, error) {
	// Read the original request body.
	var b []byte
	if r.Body != nil {
		defer func() {
			_ = r.Body.Close()
		}()

		var err error
		b, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		r.Body = io.NopCloser(bytes.NewReader(b))
	}

	// Cloning does not clone the body.
	// Set a new body for the clone.
	rc := r.Clone(r.Context())
	rc.Body = io.NopCloser(bytes.NewReader(b))

	return rc, nil
}

func FormatRequestLine(req *http.Request) string {
	reqURI := cmp.Or(req.RequestURI, req.URL.RequestURI())

	return fmt.Sprintf("%s %s HTTP/%d.%d",
		cmp.Or(req.Method, "GET"),
		reqURI,
		req.ProtoMajor,
		req.ProtoMinor,
	)
}

func NormalizeRequest(r *http.Request) {
	/*
		httputil.DumpRequest seem to strip the querystring when
		constructed with httptest.NewRequest.

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		q := r.URL.Query()
		q.Set("foo", "bar")
		r.URL.RawQuery = q.Encode()

		Output:
		GET / HTTP/1.1
		Host: example.com
	*/
	if len(r.URL.RequestURI()) > len(r.RequestURI) {
		r.RequestURI = r.URL.RequestURI()
	}
}
