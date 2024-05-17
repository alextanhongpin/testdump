package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	httpdump "github.com/alextanhongpin/dump/http"
	"github.com/google/go-cmp/cmp"
)

func TestDump(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	wr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	hd := httpdump.NewHandler(t, h)
	hd.ServeHTTP(wr, r)

	t.Run("original response is preserved", func(t *testing.T) {
		w := wr.Result()
		if w.StatusCode != http.StatusOK {
			t.Errorf("want %d, got %d", http.StatusOK, w.StatusCode)
		}
		defer w.Body.Close()
		b, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}

		want := "Hello, World!"
		got := string(bytes.TrimSpace(b))
		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})
}

func TestJSON(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		type response struct {
			Error string `json:"error"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{
			Error: "bad request",
		})
	})

	wr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte(`{"email":"john.appleseed@mail.com", "password": "12345678"}`)))
	r.Header.Set("Content-Type", "application/json")

	hd := httpdump.NewHandler(t, h)
	hd.ServeHTTP(wr, r)

	t.Run("original request is preserved", func(t *testing.T) {
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}

		want := `{"email":"john.appleseed@mail.com", "password": "12345678"}`
		got := string(bytes.TrimSpace(b))
		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})

	t.Run("original response is preserved", func(t *testing.T) {
		w := wr.Result()
		if w.StatusCode != http.StatusOK {
			t.Errorf("want %d, got %d", http.StatusOK, w.StatusCode)
		}
		defer w.Body.Close()
		b, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}

		want := `{"error":"bad request"}`
		got := string(bytes.TrimSpace(b))
		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})
}

func TestJSONCreate(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		type response struct {
			ID int `json:"id"`
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response{
			ID: rand.Intn(100),
		})
	}

	wr := httptest.NewRecorder()
	req := fmt.Sprintf(`{"createdAt": %q}`, time.Now().Format(time.RFC3339))
	r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte(req)))
	r.Header.Set("Content-Type", "application/json")

	hd := httpdump.NewHandlerFunc(t, h)
	hd.RequestComparer.Body = []cmp.Option{httpdump.IgnoreMapEntries("createdAt")}
	hd.ResponseComparer.Body = []cmp.Option{httpdump.IgnoreMapEntries("id")}
	hd.ServeHTTP(wr, r)
}

func TestMiddleware(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "hello world")
	}

	wr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Content-Type", "application/json")

	mw := []httpdump.Middleware{
		httpdump.MaskRequestHeader("[REDACTED]", "content-type"),
		httpdump.MaskResponseHeader("[REDACTED]", "content-type"),
	}
	hd := httpdump.NewHandlerFunc(t, h, mw...)
	hd.ServeHTTP(wr, r)
}

func TestMiddlewareChain(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "hello world")
	}

	wr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Content-Type", "application/json")

	mw := []httpdump.Middleware{
		// Second middleware will overwrite the first.
		httpdump.MaskRequestHeader("[ONE]", "content-type"),
		httpdump.MaskRequestHeader("[TWO]", "content-type"),
		httpdump.MaskResponseHeader("[THREE]", "content-type"),
		httpdump.MaskResponseHeader("[FOUR]", "content-type"),
	}
	hd := httpdump.NewHandlerFunc(t, h, mw...)
	hd.ServeHTTP(wr, r)

	t.Run("original request header is preserved", func(t *testing.T) {
		want := "application/json"
		got := r.Header.Get("content-type")
		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})

	t.Run("original response header is preserved", func(t *testing.T) {
		w := wr.Result()

		want := "application/json"
		got := w.Header.Get("content-type")
		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})
}

// TestHTTP shows an example of testing HTML element using goquery library.
// Suitable for handlers that returns HTMl, especially HTMX.
func TestHTTP(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<div>
	<form id='login-form'>
			<input type='email' name='email'/>
			<input type='password' name='password'/>
			<button>Submit</button>
	</form>
</div>`)
	}

	wr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Content-Type", "application/json")
	hd := httpdump.NewHandlerFunc(t, h)
	hd.ServeHTTP(wr, r)

	t.Run("checks response content-type", func(t *testing.T) {
		r := wr.Result()

		want := "text/html"
		got := r.Header.Get("Content-Type")
		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})

	t.Run("checks form exists", func(t *testing.T) {
		r := wr.Result()
		defer r.Body.Close()
		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Find the login form.
		doc.Find("#login-form").Each(func(i int, s *goquery.Selection) {
			// For each input, find the input name.
			n, ok := s.Find("input[type=email]").Attr("name")
			if !ok || n != "email" {
				t.Errorf("want %s, got %s", "email", n)
			}

			n, ok = s.Find("input[type=password]").Attr("name")
			if !ok || n != "password" {
				t.Errorf("want %s, got %s", "password", n)
			}
		})
	})
}
