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
	"net/url"
	"strings"
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

func TestQueryString(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"links": r.URL.Query(),
		})
	})

	t.Run("direct", func(t *testing.T) {
		wr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/?q=golang&limit=20", nil)

		hd := httpdump.NewHandler(t, h)
		hd.ServeHTTP(wr, r)
	})

	t.Run("programmatic", func(t *testing.T) {
		wr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		q := r.URL.Query()
		q.Add("q", "golang")
		q.Add("limit", "20")
		r.URL.RawQuery = q.Encode()

		hd := httpdump.NewHandler(t, h)
		hd.ServeHTTP(wr, r)
	})
}

func TestForm(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		json.NewEncoder(w).Encode(map[string]any{
			"form": r.Form,
		})
	})

	t.Run("no masking", func(t *testing.T) {
		wr := httptest.NewRecorder()
		formData := url.Values{
			"username": []string{"john"},
			"password": []string{"12345678"},
		}
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		hd := httpdump.NewHandler(t, h)
		hd.ServeHTTP(wr, r)
	})

	t.Run("mask field", func(t *testing.T) {
		wr := httptest.NewRecorder()
		formData := url.Values{
			"username": []string{"john"},
			"password": []string{"12345678"},
		}
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		mw := []httpdump.Middleware{
			httpdump.MaskRequestBody("[REDACTED]", "password"),
			httpdump.MaskResponseBody("[REDACTED]", "password"),
		}
		hd := httpdump.NewHandler(t, h, mw...)
		hd.ServeHTTP(wr, r)
	})
}

func TestJSON(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if w.StatusCode != http.StatusBadRequest {
			t.Errorf("want %d, got %d", http.StatusBadRequest, w.StatusCode)
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

func TestJSONDynamicFields(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
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
	hd.CompareRequestOption.Body = []cmp.Option{httpdump.IgnoreMapEntries("createdAt")}
	hd.CompareResponseOption.Body = []cmp.Option{httpdump.IgnoreMapEntries("id")}
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

// TestHTML shows an example of testing HTML element using
// goquery library.
// Suitable for handlers that returns HTMl, especially
// HTMX.
func TestHTML(t *testing.T) {
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

func TestMask(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			AccessToken string `json:"accessToken"`
			ExpiresIn   string `json:"expiresIn"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{
			AccessToken: "secret token",
			ExpiresIn:   (5 * time.Second).String(),
		})
	})

	wr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", bytes.NewReader([]byte(`{"email":"john.appleseed@mail.com", "password": "12345678"}`)))
	r.Header.Set("Content-Type", "application/json")

	mw := []httpdump.Middleware{
		httpdump.MaskRequestBody("[REDACTED]", "password"),
		httpdump.MaskResponseBody("[REDACTED]", "accessToken"),
	}
	hd := httpdump.NewHandler(t, h, mw...)
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
		if w.StatusCode != http.StatusBadRequest {
			t.Errorf("want %d, got %d", http.StatusBadRequest, w.StatusCode)
		}
		defer w.Body.Close()
		b, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatal(err)
		}

		want := `{"accessToken":"secret token","expiresIn":"5s"}`
		got := string(bytes.TrimSpace(b))
		if want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	})
}
