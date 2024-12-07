package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	fmt.Println("Connected to database")

	setupStorage()

	r := SetupRouter(conn)
	fmt.Println("Listening on port 8080")

	http.ListenAndServe(":8080", r)
}

func setupStorage() {
	_, err := os.Stat("./storage")
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir("storage", 0755)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		log.Fatal(err.Error())
	}
}
