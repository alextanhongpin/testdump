# mysqldump
[![Go Reference](https://pkg.go.dev/badge/github.com/alextanhongpin/testdump/mysqldump.svg)](https://pkg.go.dev/github.com/alextanhongpin/testdump/mysqldump)

## Purpose

`mysqldump` is a Go package designed for snapshot testing of MySQL queries. It captures query strings along with their bound arguments, enabling developers to verify and track SQL queries in their tests. This package is particularly useful for applications using ORMs or raw SQL, as it provides a way to ensure queries remain consistent over time and during migrations.

## Example Usage

```go
package main_test

import (
	"testing"
	"time"
	
	"github.com/alextanhongpin/testdump/mysqldump"
)

func TestBasicDump(t *testing.T) {
	// Create a SQL query with prepared statement arguments
	dump := &mysqldump.SQL{
		Query: "SELECT * FROM users WHERE name = ? AND age = ?",
		Args:  []any{"John", 13},
	}
	
	// Dump the query for verification
	mysqldump.Dump(t, dump)
}

func TestIgnoringDynamicValues(t *testing.T) {
	// When dealing with timestamps or other dynamic values
	dump := &mysqldump.SQL{
		Query: "SELECT * FROM users WHERE name = ? AND created_at > ?",
		Args:  []any{"John", time.Now()},
	}
	
	// Ignore the timestamp argument to prevent test flakiness
	mysqldump.Dump(t, dump, mysqldump.IgnoreArgs(":v2"))
}

func TestPrettyPrintSQL(t *testing.T) {
	dump := &mysqldump.SQL{
		Query: "SELECT * FROM users WHERE name = ? AND id = ?",
		Args:  []any{"John", 1},
	}
	
	// Create a custom dumper with pretty printing
	md := mysqldump.New(mysqldump.Prettify)
	md.Dump(t, dump)
}

func TestQueryComparison(t *testing.T) {
	// Compare queries that are semantically the same but formatted differently
	similar, err := mysqldump.CompareQuery(
		"SELECT name FROM users WHERE name = ?",
		"SELECT `name` FROM `users` WHERE name = ?"
	)
	if err != nil {
		// Handle error
	}
	if similar {
		// Queries are considered equivalent
	}
}
```

## Benefits

- **Snapshot Testing**: Create and compare SQL query snapshots to detect unintended changes
- **Query Normalization**: Compare SQL queries regardless of formatting differences, casing, or quoting styles
- **Argument Validation**: Validate prepared statement arguments in tests
- **Dynamic Value Handling**: Easily ignore timestamp fields and other dynamic values that would cause test flakiness
- **Migration Safety**: Verify queries remain functionally identical when migrating between ORMs or to raw SQL
- **Formatting Options**: Pretty-print SQL queries for better readability in test output
- **Debugging Aid**: Generate readable snapshots of executed SQL for easier debugging
- **Test Stability**: Prevent tests from breaking due to non-semantic SQL changes
- **Query Documentation**: Automatically document expected SQL queries for your codebase
