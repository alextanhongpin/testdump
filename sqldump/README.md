# sqldump

[![Go Reference](https://pkg.go.dev/badge/github.com/alextanhongpin/testdump/sqldump.svg)](https://pkg.go.dev/github.com/alextanhongpin/testdump/sqldump)

## Purpose

The `sqldump` package provides utilities for dumping SQL query results in a structured format for testing and debugging purposes. It allows developers to easily capture and validate database query outputs, making database tests more readable and maintainable.

## Example Usage

```go
package main

import (
	"context"
	"database/sql"
	"log"
	
	"github.com/alextanhongpin/testdump/sqldump"
	"github.com/alextanhongpin/testdump/yamldump"
)

func TestQuery(t *testing.T) {
	// Setup your database connection
	db, err := sql.Open("mysql", "user:password@/dbname")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	
	// Dump the results of a SQL query
	results, err := sqldump.Dump(context.Background(), db, `SELECT id, name, email FROM users`)
	if err != nil {
		t.Fatal(err)
	}
	
	// Use yamldump to format and verify the results
	// You can ignore time-based fields that might change between runs
	yamldump.Dump(t, results, yamldump.IgnoreFields("created_at"))
}
```

## Benefits

- **Simplified Testing**: Makes it easy to capture and verify database query results
- **Readable Output**: Structures query results in an easily readable format
- **Integration with yamldump**: Works seamlessly with the yamldump package for better test output formatting
- **Debugging Aid**: Helps to quickly inspect database state during development
- **Reproducible Tests**: Creates consistent test output that can be validated across test runs
- **Customization**: Allows ignoring specific fields (like timestamps) that might change between test runs
