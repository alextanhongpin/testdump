package internal

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
)

func CloneRequest(r *http.Request) (*http.Request, error) {
	// Read the original request body.
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(b))

	// Cloning does not clone the body.
	// Set a new body for the clone.
	rc := r.Clone(r.Context())
	rc.Body = io.NopCloser(bytes.NewReader(b))

	return rc, nil
}

func DumpRequest(r *http.Request, pretty bool) ([]byte, error) {
	if err := NormalizeRequest(r, pretty); err != nil {
		return nil, err
	}

	// Use `DumpRequestOut` instead of `DumpRequest` to preserve the querystring.
	// Just don't forgot to strip of the user-agent=Go-http-client/1.1 and accept-encoding=gzip
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		return nil, err
	}

	return b, nil
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

func NormalizeRequest(r *http.Request, pretty bool) error {
	// httputil.DumpRequest seem to strip the querystring
	// when constructed with httptest.NewRequest.
	if len(r.URL.RequestURI()) > len(r.RequestURI) {
		r.RequestURI = r.URL.RequestURI()
	}

	if r.Body == nil || !pretty {
		return nil
	}

	// Prettify the request body.
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	b, err = prettyBytes(b)
	if err != nil {
		return err
	}

	// httputil.DumpResponse uses different carrier line.
	// We need to standardize it.
	parts := bytes.Split(b, []byte("\n"))
	b = bytes.Join(parts, []byte("\r\n"))

	r.Body = io.NopCloser(bytes.NewReader(b))

	n := strconv.Itoa(len(b))
	if o := r.Header.Get("Content-Length"); o != n && len(b) > 0 {
		// Update the content length.
		r.Header.Set("Content-Length", n)
	}

	return nil
}
