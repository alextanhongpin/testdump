package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dumphttp "github.com/alextanhongpin/dump/http"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDump(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	wr := httptest.NewRecorder()

	dumphttp.NewHandler(t, h).ServeHTTP(wr, r)
}

func TestJSON(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		type response struct {
			Error string
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{
			Error: "bad request",
		})
	})

	r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte(`{"email":"john.appleseed@mail.com", "password": "12345678"}`)))
	r.Header.Set("Content-Type", "application/json")
	wr := httptest.NewRecorder()

	dumphttp.NewHandler(t, h).ServeHTTP(wr, r)
}

func TestJSONCreate(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		type response struct {
			ID int `json:"id"`
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{
			ID: rand.Intn(100),
		})
	})

	req := fmt.Sprintf(`{"createdAt": %q}`, time.Now().Format(time.RFC3339))
	r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte(req)))
	r.Header.Set("Content-Type", "application/json")
	wr := httptest.NewRecorder()

	dh := dumphttp.NewHandler(t, h)
	dh.RequestComparer.Body = []cmp.Option{IgnoreMapEntries("createdAt")}
	dh.ResponseComparer.Body = []cmp.Option{IgnoreMapEntries("id")}
	dh.ServeHTTP(wr, r)
}

func IgnoreMapEntries(keys ...string) cmp.Option {
	return cmpopts.IgnoreMapEntries(func(k string, v any) bool {
		for _, key := range keys {
			if key == k {
				return true
			}
		}

		return false
	})
}
