# jsondump

[![Go Reference](https://pkg.go.dev/badge/github.com/alextanhongpin/testdump/jsondump.svg)](https://pkg.go.dev/github.com/alextanhongpin/testdump/jsondump)

## Purpose

`jsondump` is a Go testing utility designed to simplify JSON assertions in tests. It dumps JSON representations of Go data structures to files that can be used for snapshot testing, making it easy to verify that your code produces the expected output.

## Example Usage

```go
package main

import (
    "testing"
    "time"

    "github.com/alextanhongpin/testdump/jsondump"
)

func TestUserData(t *testing.T) {
    // Basic usage
    user := map[string]any{
        "name": "John",
        "age": 30,
    }
    jsondump.Dump(t, user)
    
    // Ignoring time-based fields that change between test runs
    type User struct {
        Name      string    `json:"name"`
        CreatedAt time.Time `json:"createdAt"`
    }
    
    u := User{
        Name:      "John",
        CreatedAt: time.Now(),
    }
    jsondump.Dump(t, u, jsondump.IgnoreFields("createdAt"))
    
    // Masking sensitive information
    credentials := map[string]string{
        "username": "john_doe",
        "password": "secret123",
    }
    jsondump.Dump(t, credentials, jsondump.MaskFields("password"))
}
```

## Benefits

- **Simplified Testing**: No need to write lengthy assertion code - just dump the data and compare with the expected output.
- **Snapshot Testing**: Perfect for snapshot testing where you want to ensure your data structure doesn't change unexpectedly.
- **Flexible Options**:
  - Ignore dynamic values like timestamps and IDs
  - Mask sensitive fields like passwords and API keys
  - Customize output with transformers
  - Register type-specific options for consistent handling
- **Path-based Operations**: Target specific nested fields using JSONPath syntax
- **Schema Validation**: Integrates with CUE for schema validation to ensure data conforms to expected shapes
- **Type Support**: Works with any Go data structure that can be serialized to JSON

## Advanced Features

- Create custom transformers to modify output
- Apply different options to different test files
- Register type-specific options for consistent handling across tests
- Validate against CUE schemas for additional constraints
