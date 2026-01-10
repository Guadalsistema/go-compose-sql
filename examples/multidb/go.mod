module example-multidb

go 1.21

require (
	github.com/guadalsistema/go-compose-sql/v2 v2.0.0
	github.com/lib/pq v1.10.9
	modernc.org/sqlite v1.29.1
)

// Use local version for development
replace github.com/guadalsistema/go-compose-sql/v2 => ../..
