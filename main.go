//go:build lambda

package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/stinkyfingers/poopjournal/server"
	"github.com/stinkyfingers/poopjournal/storage"
)

var httpLambda *httpadapter.HandlerAdapter

func init() {
	os.Setenv("STORAGE_TYPE", "s3")

	// Initialize storage
	var store storage.Storage
	var err error

	storageType := os.Getenv("STORAGE_TYPE")
	switch storageType {
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
	case "local":
		// Fallback to local for testing
		dataDir := "/tmp/poopjournal-data"
		store, err = storage.NewLocalStorage(dataDir)
		if err != nil {
			log.Fatal("Failed to initialize local storage:", err)
		}
		log.Printf("Using local storage with data directory: %s", dataDir)
	default:
		log.Fatal("Invalid storage type for Lambda. Use 's3' or 'local'")
	}

	srv, err := server.New(store)
	if err != nil {
		log.Fatal("Failed to initialize server:", err)
	}

	mux := srv.SetupRoutes()

	// Wrap with CORS middleware
	handler := server.CorsMiddleware(mux)
	httpLambda = httpadapter.New(handler)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return httpLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
