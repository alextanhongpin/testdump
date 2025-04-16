# yamldump

[![Go Reference](https://pkg.go.dev/badge/github.com/alextanhongpin/testdump/yamldump.svg)](https://pkg.go.dev/github.com/alextanhongpin/testdump/yamldump)

## Purpose

`yamldump` is a Go testing utility that simplifies testing by allowing you to serialize test data into YAML format. It provides a convenient way to:

- Dump any Go data structure into readable YAML format
- Ignore specific fields that contain dynamic values (like timestamps)
- Mask sensitive information (like passwords and emails)
- Apply custom transformations to the output
- Work with complex nested structures using JSONPath expressions

## Example Usage

### Basic Usage

```go
import (
    "testing"
    "github.com/alextanhongpin/testdump/yamldump"
)

func TestExample(t *testing.T) {
    user := User{Name: "john", Age: 30}
    yamldump.Dump(t, user)
}
```

### Ignoring Dynamic Fields

```go
// Ignore fields by name
yamldump.Dump(t, user, yamldump.IgnoreFields("createdAt"))

// Ignore fields by path
yamldump.Dump(t, user, yamldump.IgnorePaths("$.account.createdAt"))
```

### Masking Sensitive Information

```go
// Mask sensitive fields
yamldump.Dump(t, user, yamldump.MaskFields("password", "email"))

// Use a custom mask
mask := yamldump.NewMask("REDACTED")
yamldump.Dump(t, user, mask.MaskFields("email"))

// Mask by path
yamldump.Dump(t, user, yamldump.MaskPaths("[MASKED]", []string{"$.account.email"}))
```

### Using Custom Configuration

```go
// Create a reusable dumper with configuration
yd := yamldump.New(
    yamldump.IgnorePaths("$.createdAt"),
    yamldump.MaskPaths("[REDACTED]", []string{"$.password"}),
)

// Use it multiple times
yd.Dump(t, user1)
yd.Dump(t, user2)
```

### Using Type Registry

```go
// Register type-specific handlers
yd := yamldump.New()
yd.Register(&User{}, yamldump.IgnoreFields("CreatedAt"))
yd.Register(Account{}, yamldump.IgnoreFields("UpdatedAt"))

// Then use normally
yd.Dump(t, user)
yd.Dump(t, account)
```

## Benefits

- **Improved Test Readability**: YAML format is more readable than Go structs in test outputs
- **Simplified Test Maintenance**: Easily update test expectations by viewing the YAML output
- **Flexible Data Handling**: Ignore dynamic fields or mask sensitive data
- **Handles Complex Data**: Works with nested structures, maps, slices, and custom types
- **Customizable**: Extend with your own transformers to modify the output
- **Type-aware**: Configure handlers for specific types with the registry feature
- **No Dependencies**: Lightweight with minimal external dependencies
- **Snapshot Testing**: Generate YAML snapshots for regression testing
