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
	sqlstorage "github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storage/sql"
	"github.com/Anton-Beschastnov/GoDiasoft/hw12_13_14_15_16_calendar/internal/storer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/storer_config.yaml", "Path to configuration file")
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
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		srv := &http.Server{
			Addr:              ":9093",
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
		}
		if err := srv.ListenAndServe(); err != nil {
			logg.Error("metrics server error", "error", err)
		}
	}()

	// Подключение к базе данных
	var storage app.Storage
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

	// Подключение к Kafka
	kafkaConfig := kafka.Config{
		BootstrapServers: config.Kafka.BootstrapServers,
		Topic:            config.Kafka.Topic,
	}

	// Retry с backoff для подключения к Kafka
	var consumer kafka.ConsumerInterface
	maxRetries := 30
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		consumer, err = kafka.NewConsumer(kafkaConfig)
		if err == nil {
			logg.Info("connected to kafka")
			break
		}
		logg.Warn("failed to connect to kafka, retrying", "error", err, "attempt", i+1)
		time.Sleep(retryDelay)
	}

	if consumer == nil {
		logg.Error("failed to connect to kafka after all retries")
		os.Exit(1)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			logg.Error("failed to close kafka consumer", "error", err)
		}
	}()

	// Запуск storer
	storerService := storer.New(logg, storage, consumer, storer.Config{
		KafkaTopic: config.Storer.KafkaTopic,
	})

	ctx, cancel = signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	if err := storerService.Run(ctx); err != nil {
		logg.Error("storer error", "error", err)
		os.Exit(1)
	}
}
