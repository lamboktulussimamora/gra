module hybrid-migration-demo

go 1.24.2

replace github.com/lamboktulussimamora/gra => ../../

require (
	github.com/lamboktulussimamora/gra v0.0.0-00010101000000-000000000000
	github.com/mattn/go-sqlite3 v1.14.28
)

require github.com/lib/pq v1.10.9 // indirect
