package httpdump

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/alextanhongpin/testdump/pkg/reviver"
)

type Transformer func(w *http.Response, r *http.Request) error

func maskRequestHeaders(mask string, fields ...string) Transformer {
	return func(w *http.Response, r *http.Request) error {
		for _, field := range fields {
			v := r.Header.Get(field)
			if v == "" {
				return fmt.Errorf("missing header %s", field)
			}

			r.Header.Set(field, mask)
		}

		return nil
	}
}

func maskResponseHeaders(mask string, fields ...string) Transformer {
	return func(w *http.Response, r *http.Request) error {
		for _, field := range fields {
			v := w.Header.Get(field)
			if v == "" {
				return fmt.Errorf("missing header %s", field)
			}

			w.Header.Set(field, mask)
		}

		return nil
	}
}

func maskRequestFields(mask string, fields ...string) Transformer {
	return func(w *http.Response, r *http.Request) error {
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		// This is only available for request, since only
		// request can contain form data.
		if !json.Valid(b) {
			// Could this be a form data?
			v, err := url.ParseQuery(string(b))
			if err != nil {
				// Skip processing.
				r.Body = io.NopCloser(bytes.NewReader(b))
				return nil
			}
			for _, f := range fields {
				if v.Get(f) == "" {
					return fmt.Errorf("missing field %s", f)
				}
				v.Set(f, mask)
			}
			r.Body = io.NopCloser(strings.NewReader(v.Encode()))
			return nil
		}

		var t map[string]any
		if err := reviver.Unmarshal(b, &t, func(key string, val any) (any, error) {
			path := reviver.Base(key)
			for _, f := range fields {
				if f == path {
					// Only mask the value if it is a string.
					if _, ok := val.(string); ok {
						return mask, nil
					}
				}
			}

			return val, nil
		}); err != nil {
			return err
		}

		b, err = json.Marshal(t)
		if err != nil {
			return err
		}
		r.Body = io.NopCloser(bytes.NewReader(b))

		return nil
	}
}

func maskResponseFields(mask string, fields ...string) Transformer {
	return func(w *http.Response, r *http.Request) error {
		defer w.Body.Close()

		b, err := io.ReadAll(w.Body)
		if err != nil {
			return err
		}
		if !json.Valid(b) {
			w.Body = io.NopCloser(bytes.NewReader(b))
			return nil
		}

		var t map[string]any
		if err := reviver.Unmarshal(b, &t, func(key string, val any) (any, error) {
			path := reviver.Base(key)
			for _, f := range fields {
				if f == path {
					// Only mask the value if it is a string.
					if _, ok := val.(string); ok {
						return mask, nil
					}
				}
			}

			return val, nil
		}); err != nil {
			return err
		}

		b, err = json.Marshal(t)
		if err != nil {
			return err
		}
		w.Body = io.NopCloser(bytes.NewReader(b))

		return nil
	}
}
