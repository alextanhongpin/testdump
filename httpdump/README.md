# httpdump

[![Go Reference](https://pkg.go.dev/badge/github.com/alextanhongpin/testdump/httpdump.svg)](https://pkg.go.dev/github.com/alextanhongpin/testdump/httpdump)

## Installation

To use this package, you can import it like so:

```go
import "github.com/alextanhongpin/testdump/httpdump"
```

## Overview

The `dump/http` package provides utilities for dumping HTTP requests and responses in Go. It includes functionality for capturing, comparing, and storing HTTP request/response snapshots, which can be useful for snapshot testing.


## Usage

### Creating a New Handler

You can create a new handler using the `NewHandler` function. This function takes a `testing.T` instance, an `http.Handler`, and an optional list of transformers. Here's an example:

```go
h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte("Hello, World!"))
})

hd := httpdump.Handler(t, h)
```

### Serving HTTP Requests
You can use the ServeHTTP method of the Handler struct to serve HTTP requests. Here's an example:

```go
wr := httptest.NewRecorder()
r := httptest.NewRequest(http.MethodGet, "/", nil)

hd.ServeHTTP(wr, r)
```


### Testing the Response

The `httpdump` package does not modify the original request or response payload.
You can test the response of the handler using the Result method of the ResponseRecorder and the ReadAll function from the io package. Here's an example:


```go
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
```

### Testdata Snapshot

The `dump/http` package also provides a convenient way to generate example snapshots for testing purposes. In the `testdata` directory of the package, you can find pre-generated snapshots that represent expected HTTP request/response pairs. These snapshots can be used as a reference to prevent regression in your code.

Here is an example from the [testdata/TestDump.http](testdata/TestDump.http)

```http
-- request.http --
GET / HTTP/1.1
Host: example.com

-- response.http --
HTTP/1.1 200 OK
Connection: close

Hello, World!

```

### Transformers

Transformers allows modifying the request/response body before snapshotting. 
You can for example mask headers that contains random or sensitive values.

```go
mw := []httpdump.Transformer{
  // Second middleware will overwrite the first.
  httpdump.MaskRequestHeader("[REDACTED]", "Date"),
  httpdump.MaskResponseHeader("[REDACTED]", "Date"),
  httpdump.MaskRequestBody("[REDACTED]", "password"),
  httpdump.MaskResponseBody("[REDACTED]", "accessToken"),
}
hd := httpdump.NewHandlerFunc(t, h, mw...)
hd.ServeHTTP(wr, r)
```

### Ignoring JSON fields from comparison

If the JSON request contains dynamic values such as uuid or datetime, you can skip those fields from comparison:

```go
hd := httpdump.HandlerFunc(t, h,
  httpdump.IgnoreRequestFields("createdAt"),
  httpdump.IgnoreResponseFields("id"),
)
hd.ServeHTTP(wr, r)
```

### Diff

When the content of the generated dump doesn't match the snapshot, you can see the diff error.

```bash
➜  http git:(http) ✗ gotest
--- FAIL: TestDump (0.00s)
    http_test.go:30: Response Body: 
        
          Snapshot(-)
          Received(+)
        
        
          string(
        -       "Hello, world!",
        +       "Hello, World!",
          )
        
FAIL
exit status 1
FAIL    github.com/alextanhongpin/testdump/http     0.449s
```
