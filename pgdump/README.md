# pgdump
[![Go Reference](https://pkg.go.dev/badge/github.com/alextanhongpin/testdump/pgdump.svg)](https://pkg.go.dev/github.com/alextanhongpin/testdump/pgdump)


`pgdump` and it's cousin `mysqldump` allows you to perform snapshot testing for sql.

It dumps the query string with the arguments extracted, and displays the diff whenever there are changes.

If you are using an ORM, this will save you a lot in the long run, especially when migrating to a different ORM or even using raw sql.

It improves observability and allows you to verify the generated SQL. Different ORM generates SQL differently, for example different casing, quoted tables and/or column definitions. Even handwritten ones might differ, but this library takes that into consideration when comparing the queries. In short, when switching to another ORM, your tests will not break just because one ORM prefers quoting a column etc.

Using `pgdump.NewRecorder`, you can generate snapshots of all the executed SQL in a run. One advantage is you can easily copy the generated SQL to run in your clients for easier debugging.
