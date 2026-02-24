package handler

import (
	"context"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maxbrt/steamzip-app/internal/database"
	"github.com/maxbrt/steamzip-app/internal/handlers"
	"github.com/maxbrt/steamzip-app/internal/server"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

var (
	once        sync.Once
	httpHandler http.Handler
)

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		db := database.Connect()

		awsCfg, _ := config.LoadDefaultConfig(context.Background())
		s3Client := s3.NewFromConfig(awsCfg)

		bucket              := utils.GetEnv("S3_BUCKET", "")
		stripeSecretKey     := utils.GetEnv("STRIPE_SECRET_KEY", "")
		stripeWebhookSecret := utils.GetEnv("STRIPE_WEBHOOK_SECRET", "")
		frontendURL         := utils.GetEnv("FRONTEND_URL", "")

		s := server.NewServer()
		h := handlers.NewHandler(db, s3Client, bucket, stripeSecretKey, stripeWebhookSecret, frontendURL)
		s.MountHandlers(h)
		httpHandler = s.Router
	})
	httpHandler.ServeHTTP(w, r)
}
