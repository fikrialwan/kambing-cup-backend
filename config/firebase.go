package config

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"google.golang.org/api/option"
)

// SetupFirebase initializes the Firebase application and returns the Realtime Database client.
// It requires "serviceAccountKey.json" to be present in the root directory
// and FIREBASE_DATABASE_URL environment variable to be set.
func SetupFirebase() *db.Client {
	ctx := context.Background()

	serviceAccountKeyPath := "serviceAccountKey.json"

	opt := option.WithCredentialsFile(serviceAccountKeyPath)

	conf := &firebase.Config{
		DatabaseURL: os.Getenv("FIREBASE_DATABASE_URL"),
	}

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	client, err := app.Database(ctx)
	if err != nil {
		log.Fatalf("error initializing firebase database client: %v", err)
	}

	log.Println("Connected to Firebase Realtime Database")
	return client
}
