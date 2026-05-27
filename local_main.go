//go:build !lambda

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/stinkyfingers/poopjournal/server"
	"github.com/stinkyfingers/poopjournal/storage"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Initialize storage based on environment
	var store storage.Storage
	var err error

	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	switch storageType {
	case "local":
		dataDir := os.Getenv("DATA_DIR")
		if dataDir == "" {
			dataDir = "./data"
		}
		store, err = storage.NewLocalStorage(dataDir)
		if err != nil {
			log.Fatal("Failed to initialize local storage:", err)
		}
		log.Printf("Using local storage with data directory: %s", dataDir)
	case "s3":
		bucketName := os.Getenv("S3_BUCKET")
		if bucketName == "" {
			log.Fatal("S3_BUCKET environment variable is required")
		}

		store, err = storage.NewS3Storage(bucketName)
		if err != nil {
			log.Fatal("Failed to initialize S3 storage:", err)
		}
		log.Printf("Using S3 storage with bucket: %s", bucketName)
	default:
		log.Fatal("Invalid storage type. Use 'local' or 's3'")
	}

	srv, err := server.New(store)
	if err != nil {
		log.Fatal("Failed to initialize server:", err)
	}

	mux := srv.SetupRoutes()

	// Wrap with CORS middleware
	handler := server.CorsMiddleware(mux)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8070"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	log.Printf("Visit http://localhost:%s to access the application", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
