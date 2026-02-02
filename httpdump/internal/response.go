package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
)

func CloneResponse(w *http.Response) (*http.Response, error) {
	// There are no way to clone the response.
	// So we just dump and re-read it again.
	b, err := httputil.DumpResponse(w, true)
	if err != nil {
		return nil, err
	}

	bb := bufio.NewReader(bytes.NewReader(b))
	wc, err := http.ReadResponse(bb, nil)
	return wc, err
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
