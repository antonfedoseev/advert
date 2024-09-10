package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"internal/constant"
	"internal/global"
	"internal/rpc"
	"internal/settings"
	"internal/upload"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path"
	"pkg/expath"
	"pkg/mb"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	logo = `
      ||
     ||||
    ||  ||
   ||    ||
  ||||||||||
 ||        ||
||          ||
`
	appName = "advertd"
)

var (
	pidFile = flag.String("pid-file", "advertd.pid", "path to to pid file")
)

func main() {
	ctx, quit := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGHUP)
	defer quit()

	exPath, err := expath.Get()
	if err != nil {
		panic("failed to get executable path: " + err.Error())
	}

	settings, err := initSettings(path.Join(exPath, "settings.json"))
	if err != nil {
		panic("failed to read settings: " + err.Error())
	}

	logger := newLogger(settings.LogLevel, constant.LogAppPrefix)
	logger.Info("Starting advertd", "executable_path", exPath,
		"version", constant.AppVersion, "runtime_version", runtime.Version(), "os", runtime.GOOS, "arch", runtime.GOARCH,
		"addr", settings.UrlListen, "pid_file", *pidFile)
	logger.V(1).Info(logo)

	producer := mb.NewProducer(settings.MessageBroker)

	hub := global.New(exPath, settings, logger, appName, producer)
	defer hub.Dispose()

	consumer := mb.NewConsumer(ctx, settings.MessageBroker, logger, rpc.NewMbHandler(hub))
	defer consumer.Close()

	startPprof()

	srv := newServer(hub)

	<-ctx.Done()

	// Make sure to set a deadline on exiting the process
	// after upg.Exit() is closed. No new upgrades can be
	// performed if the parent doesn't exit.
	time.AfterFunc(30*time.Second, func() {
		logger.Error(nil, "Graceful shutdown timed out")
		os.Exit(1)
	})

	// Wait for connections to drain.
	srv.Shutdown(context.Background())

	logger.Info("Exiting", "pid", os.Getpid())
}

func initSettings(path string) (settings.Settings, error) {
	s := settings.Settings{}
	err := s.Read(path)
	return s, err
}

func newLogger(logLevel int, app string) logr.Logger {
	var logger logr.Logger
	//setting production level
	if logLevel == 0 {
		core := zapcore.NewTee(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.Lock(os.Stdout),
				zap.LevelEnablerFunc(func(level zapcore.Level) bool { return level == zapcore.InfoLevel }),
			),
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.Lock(os.Stderr),
				zap.LevelEnablerFunc(func(level zapcore.Level) bool { return level == zapcore.ErrorLevel }),
			),
		)
		logger = zapr.NewLogger(zap.New(core))
	} else {
		colored := strings.Contains(os.Getenv("ANSI_COLOR"), "1")
		zc := zap.NewDevelopmentConfig()
		zc.EncoderConfig.ConsoleSeparator = " "
		if colored {
			zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		zl, _ := zc.Build()
		logger = zapr.NewLogger(zl)
	}

	logger = logger.WithValues("pid", syscall.Getpid(), "app", app)

	return logger
}

func startPprof() {
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "")
		})
		log.Fatal(http.ListenAndServe(":7200", nil))
	}()
}

func newServer(globs global.Hub) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}

	ln, err := net.Listen("tcp", globs.Settings.UrlListen)
	if err != nil {
		globs.Logger.Error(err, "Can't listen")
		os.Exit(1)
	}

	if err := initHandlers(mux, globs); err != nil {
		globs.Logger.Error(err, "Error initializing server")
		os.Exit(1)
	}

	go func() {
		if err := srv.Serve(ln); err != nil {
			globs.Logger.Error(err, "Error running server")
			os.Exit(1)
		}
	}()

	return srv
}

func initHandlers(mux *http.ServeMux, globs global.Hub) error {
	mux.Handle("/gateway_create_advert", upload.NewServer(globs))

	return nil
}
