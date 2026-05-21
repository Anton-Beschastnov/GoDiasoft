//go:build !duckdb
// +build !duckdb

package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/pressly/goose/v3"
)

func main() {
	dir := flag.String("dir", "migrations", "directory with migration files")
	dbURL := flag.String("dburl", "", "database URL")
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("dburl is required")
	}

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Используем os.DirFS для работы с файловой системой
	goose.SetBaseFS(os.DirFS(*dir))
	if err := goose.Run("up", db, *dir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
}
