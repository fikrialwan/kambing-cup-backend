package main

import (
	"context"
	"fmt"
	"kambing-cup-backend/config"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	fmt.Println("Connected to database")

	runDatabaseMigrations(os.Getenv("DATABASE_URL"))

	config.SetupStorage()
	r := config.SetupRouter(conn)

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", r)
}

func runDatabaseMigrations(dbURL string) {
    m, err := migrate.New("file://migrations", dbURL)
    if err != nil {
        log.Fatalf("Could not create migrate instance: %v", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatalf("Could not run up migrations: %v", err)
    }

    log.Println("Migrations completed successfully!")
}