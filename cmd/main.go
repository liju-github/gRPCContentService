package main

import (
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"github.com/liju-github/ContentService/internal/models"
	"github.com/liju-github/ContentService/internal/repository"
	"github.com/liju-github/ContentService/internal/service"
	contentPB "github.com/liju-github/ContentService/proto/content"
)

func main() {
    if err := godotenv.Load(".env"); err != nil {
        log.Fatal("Error loading .env file")
    }

    mongoURI := os.Getenv("MONGO_URI")
    dbName := os.Getenv("MONGO_DB_NAME")
    cfg := models.MongoConfig{
        URI: mongoURI,
        Database: dbName,
    }
    port := os.Getenv("PORT")

    repo, err := mongodb.NewMongoRepository(&cfg)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    contentService := service.NewContentService(repo)

    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    server := grpc.NewServer()
    contentPB.RegisterContentServiceServer(server, contentService)

    log.Printf("Server listening on port %s", port)
    if err := server.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}