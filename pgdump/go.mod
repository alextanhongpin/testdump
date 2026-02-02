module github.com/alextanhongpin/testdump/pgdump

go 1.24.0

toolchain go1.24.2

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/alextanhongpin/testdump/pkg/diff v0.0.0-20260202055853-a19b226ed7bf
	github.com/alextanhongpin/testdump/pkg/file v0.0.0-20260202055853-a19b226ed7bf
	github.com/alextanhongpin/testdump/pkg/snapshot v0.0.0-20260202055853-a19b226ed7bf
	github.com/alextanhongpin/testdump/pkg/sqlformat v0.0.0-20260202055853-a19b226ed7bf
	github.com/google/go-cmp v0.7.0
	github.com/pganalyze/pg_query_go/v4 v4.2.3
	github.com/pganalyze/pg_query_go/v6 v6.2.2
	golang.org/x/tools v0.41.0
)

require (
	github.com/golang/protobuf v1.5.4 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
