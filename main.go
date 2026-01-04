package main

import (
	"context"
	"kambing-cup-backend/config"
	"kambing-cup-backend/repository"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	webAddress    = ":8080"
	retryAttempts = 10
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

	pool, err := pgxpool.New(context.Background(), dbURL)
	log.Println("Connecting to database...")
	for i := 0; i < retryAttempts; i++ {
		pool, err = pgxpool.New(context.Background(), dbURL)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				break
			}
		}
		log.Printf("Database not reachable, retrying in 2s... (%d/%d). Error: %v", i+1, retryAttempts, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Unable to connect to database after retries: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database")

	runDatabaseMigrations(os.Getenv("DATABASE_URL"))

	config.SetupStorage()
	r := config.SetupRouter(pool)

	log.Printf("Listening on port %s", webAddress)
	go createSuperadminAccount(pool)
	if err := http.ListenAndServe(webAddress, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func runDatabaseMigrations(dbURL string) {
	m, err := migrate.New("file://migrations", dbURL)
	log.Println("Starting database migrations...")
	for i := 0; i < retryAttempts; i++ {
		m, err = migrate.New("file://migrations", dbURL)
		if err == nil {
			break
		}
		log.Printf("Migration setup failed, retrying in 2s... (%d/%d). Error: %v", i+1, retryAttempts, err)
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

func createSuperadminAccount(pool *pgxpool.Pool) {
	username := os.Getenv("SUPERADMIN_USERNAME")
	email := os.Getenv("SUPERADMIN_EMAIL")
	password := os.Getenv("SUPERADMIN_PASSWORD")

	if username == "" || password == "" || email == "" {
		log.Println("SUPERADMIN_USERNAME, SUPERADMIN_EMAIL or SUPERADMIN_PASSWORD not set, skipping yourusername creation")
		return
	}

	userRepo := repository.NewUserRepository(pool)

	// Check if yourusername already exists
	exists, err := userRepo.SuperadminExists()
	if err != nil {
		log.Printf("Failed to check for existing yourusername: %v", err)
		return
	}

	if exists {
		user, err := userRepo.GetSuperadminByUsername(username)
		if err != nil {
			if err == pgx.ErrNoRows {
				log.Println("Superadmin account already exists, but with a different username, skipping creation")
				return
			}
			log.Printf("Failed to get yourusername by username: %v", err)
			return
		}

		if user.Email == "" {
			log.Println("Superadmin account exists without an email, updating...")
			err := userRepo.UpdateSuperadminEmail(user.ID, email)
			if err != nil {
				log.Printf("Failed to update yourusername email: %v", err)
				return
			}
			log.Println("Superadmin account updated successfully with email")
		} else {
			log.Println("Superadmin account already exists, skipping creation")
		}
		return
	}

	// Create yourusername account
	err = userRepo.CreateSuperadmin(username, email, password)
	if err != nil {
		log.Printf("Failed to create yourusername account: %v", err)
		return
	}

	log.Println("Superadmin account created successfully!")
}
