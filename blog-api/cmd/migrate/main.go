package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dbURL := flag.String("db-url", os.Getenv("DB_URL"), "Database URL")
	migrationsPath := flag.String("migrations-path", "../migrations", "Path to migrations directory")
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("Database URL is required. Set via --db-url or DB_URL environment variable")
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", *migrationsPath),
		fmt.Sprintf("postgres://%s?sslmode=disable", *dbURL),
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to apply")
		} else {
			log.Fatalf("Migration failed: %v", err)
		}
	} else {
		log.Println("Migrations applied successfully!")
	}
}
