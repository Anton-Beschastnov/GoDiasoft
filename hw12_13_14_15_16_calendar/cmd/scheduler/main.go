package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/kafka"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/scheduler"
	memorystorage "github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/scheduler_config.yaml", "Path to configuration file")
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

	// Metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9091", nil); err != nil {
			logg.Error("metrics server error", "error", err)
		}
	}()

	// Подключение к базе данных
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

	// Подключение к Kafka
	kafkaConfig := kafka.Config{
		BootstrapServers: config.Kafka.BootstrapServers,
		Topic:            config.Kafka.Topic,
	}

	// Retry с backoff для подключения к Kafka
	var producer kafka.ProducerInterface
	maxRetries := 30
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		producer, err = kafka.NewProducer(kafkaConfig)
		if err == nil {
			logg.Info("connected to kafka")
			break
		}
		logg.Warn("failed to connect to kafka, retrying", "error", err, "attempt", i+1)
		time.Sleep(retryDelay)
	}

	if producer == nil {
		logg.Error("failed to connect to kafka after all retries")
		os.Exit(1)
	}

	defer func() {
		if err := producer.Close(); err != nil {
			logg.Error("failed to close kafka producer", "error", err)
		}
	}()

	// Запуск scheduler
	sched := scheduler.New(logg, storage, producer, config.Scheduler)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	logg.Info("scheduler is running...")

	if err := sched.Run(ctx); err != nil {
		logg.Error("scheduler error", "error", err)
		os.Exit(1)
	}
}
