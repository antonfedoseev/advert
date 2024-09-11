module advertd

go 1.22.3

require (
	go.uber.org/automaxprocs v1.5.3
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	go.uber.org/zap v1.19.0
	internal v0.0.0
	pkg v0.0.0
)

require (
	github.com/go-redsync/redsync/v4 v4.13.0 // indirect
	github.com/gomodule/redigo v1.9.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/huandu/go-sqlbuilder v1.28.0 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
)

replace internal => ../internal/

replace pkg => ../pkg/
