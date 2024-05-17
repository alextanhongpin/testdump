package http

import (
	"fmt"
	"net/http"
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
