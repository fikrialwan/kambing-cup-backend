package main

import (
	"context"
	"fmt"
	"kambing-cup-backend/config"
	"kambing-cup-backend/repository"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	log.Println("Connecting to database...")
	for i := 0; i < 10; i++ {
		conn, err = pgx.Connect(context.Background(), dbURL)
		if err == nil {
			break
		}
		log.Printf("Database not reachable, retrying in 2s... (%d/10). Error: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Unable to connect to database after retries: %v", err)
	}
	defer conn.Close(context.Background())

	fmt.Println("Connected to database")

	runDatabaseMigrations(os.Getenv("DATABASE_URL"))

	config.SetupStorage()
	r := config.SetupRouter(conn)

	fmt.Println("Listening on port 8080")
	go createSuperadminAccount(conn)
	http.ListenAndServe(":8080", r)
}

func runDatabaseMigrations(dbURL string) {
	m, err := migrate.New("file://migrations", dbURL)
	log.Println("Starting database migrations...")
	for i := 0; i < 10; i++ {
		m, err = migrate.New("file://migrations", dbURL)
		if err == nil {
			break
		}
		log.Printf("Migration setup failed, retrying in 2s... (%d/10). Error: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Could not run up migrations: %v", err)
	}

	log.Println("Migrations completed successfully!")
}

func createSuperadminAccount(conn *pgx.Conn) {
	username := os.Getenv("SUPERADMIN_USERNAME")
	password := os.Getenv("SUPERADMIN_PASSWORD")

	if username == "" || password == "" {
		log.Println("SUPERADMIN_USERNAME or SUPERADMIN_PASSWORD not set, skipping yourusername creation")
		return
	}

	userRepo := repository.NewUserRepository(conn)

	// Check if yourusername already exists
	exists, err := userRepo.SuperadminExists()
	if err != nil {
		log.Printf("Failed to check for existing yourusername: %v", err)
		return
	}

	if exists {
		log.Println("Superadmin account already exists, skipping creation")
		return
	}

	// Create yourusername account
	err = userRepo.CreateSuperadmin(username, password)
	if err != nil {
		log.Printf("Failed to create yourusername account: %v", err)
		return
	}

	log.Println("Superadmin account created successfully!")
}
