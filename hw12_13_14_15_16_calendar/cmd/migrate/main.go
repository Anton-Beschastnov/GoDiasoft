package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dbURL := flag.String("dburl", "", "database URL")
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("dburl is required")
	}

	db, err := sql.Open("pgx", *dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Определяем путь к миграциям относительно места запуска бинарника
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)
	migrationsDir := filepath.Join(execDir, "migrations")

	// Проверяем существование директории миграций
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Fatalf("migrations directory does not exist: %s", migrationsDir)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	// goose требует только .up.sql файлы в директории, .down.sql файлы игнорируются
	// при запуске миграций вверх
	if err := goose.Run("up", db, migrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	fmt.Println("migrations completed successfully")
}
