package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alextanhongpin/dump/pkg/reviver"
)

type Middleware func(w *http.Response, r *http.Request) error

func MaskRequestHeader(mask string, fields ...string) Middleware {
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

func MaskResponseHeader(mask string, fields ...string) Middleware {
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

func MaskRequestBody(mask string, fields ...string) Middleware {
	return func(w *http.Response, r *http.Request) error {
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		var m map[string]any
		if err := reviver.Unmarshal(b, &m, func(key string, val any) (any, error) {
			path := reviver.Base(key)
			for _, f := range fields {
				if f == path {
					return mask, nil
				}
			}

			return val, nil
		}); err != nil {
			return err
		}

		b, err = json.Marshal(m)
		if err != nil {
			return err
		}
		r.Body = io.NopCloser(bytes.NewReader(b))

		return nil
	}
}

func MaskResponseBody(mask string, fields ...string) Middleware {
	return func(w *http.Response, r *http.Request) error {
		defer w.Body.Close()

		b, err := io.ReadAll(w.Body)
		if err != nil {
			return err
		}

		var m map[string]any
		if err := reviver.Unmarshal(b, &m, func(key string, val any) (any, error) {
			path := reviver.Base(key)
			for _, f := range fields {
				if f == path {
					return mask, nil
				}
			}

			return val, nil
		}); err != nil {
			return err
		}

		b, err = json.Marshal(m)
		if err != nil {
			return err
		}
		w.Body = io.NopCloser(bytes.NewReader(b))

		return nil
	}
}
