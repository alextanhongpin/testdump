# testdump

> Snapshot testing for Go
 
`testdump` is a simple Go package for snapshot testing. It helps you ensure that your code produces the same output over time.

Dumps for different format (`json`, `yaml`, `pg`, `mysql`, `sql`, `grpc`, `http`, `text`) is available in the respective directory.

See the tests for more example.

## Basic usage

```go
package main

import (
	"testing"
	"time"

	"github.com/alextanhongpin/testdump/jsondump"
)

type User struct {
	Age       int
	Name      string
	CreatedAt time.Time
}

func TestJSONDump(t *testing.T) {
	jsondump.Dump(t, &User{
		Age:       13,
		Name:      "John Appleseed",
		CreatedAt: time.Now(),
	})
}
```

Run this locally `go test -v`.

It will produce a dump in `testdata/TestJSONDump/main.User.json`:

```json
{
 "Age": 13,
 "Name": "John Appleseed",
 "CreatedAt": "2024-06-19T02:08:54.802509+08:00"
}
```

Make any changes to the values, or just re-run it again to see this error:

```bash
➜  go-test-jsondump go test -v
=== RUN   TestJSONDump
    json.go:28:

          Snapshot(-)
          Received(+)


          map[string]any{
                "Age": float64(13),
                "CreatedAt": strings.Join({
                        "2024-06-19T02:",
        -               "08:54.802509",
        +               "10:05.119235",
                        "+08:00",
                }, ""),
                "Name": string("John Appleseed"),
          }

--- FAIL: TestJSONDump (0.00s)
FAIL
exit status 1
FAIL    github.com/alextanhongpin/go-test-jsondump      0.243s
```

Whenever there are difference in the generated snapshot, testdump will show the diffs. In this scenario, the `CreatedAt` value changes every time the test is run.

For such cases, we can choose to ignore the fields from comparison. Update the test to ignore the field `CreatedAt`:

```diff
func TestJSONDump(t *testing.T) {
	jsondump.Dump(t, &User{
		Age:       13,
		Name:      "John Appleseed",
		CreatedAt: time.Now(),
+	}, jsondump.IgnoreFields("CreatedAt"))
}
```

Re-running this will now result in a successful test:

```bash
➜  go-test-jsondump go test -v
=== RUN   TestJSONDump
--- PASS: TestJSONDump (0.00s)
PASS
ok      github.com/alextanhongpin/go-test-jsondump      0.551s
```
