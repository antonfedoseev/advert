module internal

go 1.22.3

require (
	github.com/jmoiron/sqlx v1.4.0
	github.com/go-logr/logr v1.2.3
	github.com/go-redsync/redsync/v4 v4.13.0
	github.com/gomodule/redigo v1.9.2
	github.com/pkg/errors v0.9.1
	github.com/segmentio/kafka-go v0.4.47
	golang.org/x/sync v0.3.0
	pkg v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)

replace pkg => ../pkg
