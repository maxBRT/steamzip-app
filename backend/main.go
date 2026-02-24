package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maxbrt/steamzip-app/internal/database"
	"github.com/maxbrt/steamzip-app/internal/handlers"
	"github.com/maxbrt/steamzip-app/internal/server"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

func main() {
	db := database.Connect()
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Auto-migration failed: %v", err)
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	bucket := utils.GetEnv("S3_BUCKET", "")
	stripeSecretKey := utils.GetEnv("STRIPE_SECRET_KEY", "")
	stripeWebhookSecret := utils.GetEnv("STRIPE_WEBHOOK_SECRET", "")
	frontendURL := utils.GetEnv("FRONTEND_URL", "")

	s := server.NewServer()
	h := handlers.NewHandler(db, s3Client, bucket, stripeSecretKey, stripeWebhookSecret, frontendURL)
	s.MountHandlers(h)

	port := utils.GetEnv("PORT", "3000")
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, s.Router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
