package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

func CloneResponse(w *http.Response) (*http.Response, error) {
	// There are no way to clone the response.
	// So we just dump and re-read it again.
	b, err := DumpResponse(w, false)
	if err != nil {
		return nil, err
	}

	bb := bufio.NewReader(bytes.NewReader(b))
	wc, err := http.ReadResponse(bb, nil)
	return wc, err
}

func DumpResponse(w *http.Response, pretty bool) ([]byte, error) {
	// Before dumping
	if err := NormalizeResponse(w, pretty); err != nil {
		return nil, err
	}

	b, err := httputil.DumpResponse(w, true)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func NormalizeResponse(w *http.Response, pretty bool) error {
	if !pretty {
		return nil
	}

	// Prettify the request body.
	b, err := io.ReadAll(w.Body)
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

	w.Body = io.NopCloser(bytes.NewReader(b))

	// If the content-length is set, we need to update it.
	n := int64(len(b))
	if w.ContentLength > 0 && w.ContentLength != n {
		w.ContentLength = n
	}

	return nil
}

func FormatResponseLine(w *http.Response) string {
	// Status line
	text := w.Status
	if text == "" {
		text = http.StatusText(w.StatusCode)
		if text == "" {
			text = "status code " + strconv.Itoa(w.StatusCode)
		}
	} else {
		// Just to reduce stutter, if user set w.Status to "200 OK" and StatusCode to 200.
		// Not important.
		text = strings.TrimPrefix(text, strconv.Itoa(w.StatusCode)+" ")
	}

	return fmt.Sprintf("HTTP/%d.%d %03d %s",
		w.ProtoMajor,
		w.ProtoMinor,
		w.StatusCode,
		text,
	)
}
