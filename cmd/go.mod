module advertd

go 1.22.3

require (
	github.com/emirpasic/gods v1.18.1
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/zapr v1.2.3
	github.com/segmentio/kafka-go v0.4.47
	go.uber.org/zap v1.19.0
	internal v0.0.0
	pkg v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
)

replace internal => ../internal/

replace pkg => ../pkg/
