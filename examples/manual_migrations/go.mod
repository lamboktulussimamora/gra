module github.com/lamboktulussimamora/gra/examples/manual_migrations

go 1.21

require github.com/lib/pq v1.10.9

replace github.com/lamboktulussimamora/gra/examples/manual_migrations/tools/common => ./tools/common

replace github.com/lamboktulussimamora/gra/examples/manual_migrations/schema => ./schema
