package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/server/http"
	memorystorage "github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
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

		// Apply migrations
		if err := runMigrations(dsn); err != nil {
			logg.Error("failed to run migrations", "error", err)
			os.Exit(1)
		}

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

	handler := internalhttp.NewCalendarHandler(logg, storage)
	server := internalhttp.NewServer(logg, handler, config.HTTP.Host, config.HTTP.Port)

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

func runMigrations(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Путь к миграциям - определяем относительно исполняемого файла
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	migrationsDir := filepath.Join(filepath.Dir(exePath), "migrations")

	// Проверяем существование директории
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", migrationsDir)
	}

	ctx := context.Background()
	if err := goose.RunContext(ctx, "up", db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
