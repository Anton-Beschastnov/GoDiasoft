package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/internal/app"
	"github.com/Anton-Beschastnov/GoDiasoft/internal/logger"
	internalhttp "github.com/Anton-Beschastnov/GoDiasoft/internal/server/http"
	memorystorage "github.com/Anton-Beschastnov/GoDiasoft/internal/storage/memory"
	sqlstorage "github.com/Anton-Beschastnov/GoDiasoft/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/config.yaml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logg := logger.New(config.Logger.Level, os.Stdout)

	var storage app.Storage
	switch config.Storage.Type {
	case "sql":
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			config.DB.User,
			config.DB.Password,
			config.DB.Host,
			config.DB.Port,
			config.DB.Database,
		)
		sqlStore := sqlstorage.New(dsn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := sqlStore.Connect(ctx); err != nil {
			logg.Error("failed to connect to database", "error", err)
			cancel()
			os.Exit(1)
		}
		cancel()
		storage = sqlStore
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := sqlStore.Close(ctx); err != nil {
				logg.Error("failed to close database connection", "error", err)
			}
		}()
	default:
		storage = memorystorage.New()
	}

	calendar := app.New(logg, storage)
	server := internalhttp.NewServer(logg, calendar, config.HTTP.Host, config.HTTP.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			logg.Error("failed to stop http server", "error", err)
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server", "error", err)
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
