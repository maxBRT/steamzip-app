package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maxbrt/steamzip-app/internal/models"
	"gorm.io/gorm"
)

type Handler struct {
	Database            *gorm.DB
	S3Client            *s3.Client
	S3PresignClient     *s3.PresignClient
	S3Bucket            string
	StripeSecretKey     string
	StripeWebhookSecret string
	FrontendURL         string
}

func NewHandler(db *gorm.DB, s3Client *s3.Client, bucket, stripeSecretKey, stripeWebhookSecret, frontendURL string) *Handler {
	return &Handler{
		Database:            db,
		S3Client:            s3Client,
		S3PresignClient:     s3.NewPresignClient(s3Client),
		S3Bucket:            bucket,
		StripeSecretKey:     stripeSecretKey,
		StripeWebhookSecret: stripeWebhookSecret,
		FrontendURL:         frontendURL,
	}
}

// loadSession parses the sessionId URL param, loads the session and its most
// recent asset from the database, and checks for expiry. If anything fails it
// writes the appropriate HTTP error response and returns (nil, false). The
// caller must return immediately when ok is false.
func (h *Handler) loadSession(w http.ResponseWriter, r *http.Request) (*models.Session, bool) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return nil, false
	}

	session := models.Session{}
	if err := h.Database.Preload("Asset", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).First(&session, sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Session not found", http.StatusNotFound)
			return nil, false
		}
		http.Error(w, "Error loading session", http.StatusInternalServerError)
		return nil, false
	}

	if session.ExpiresAt.Before(time.Now()) {
		http.Error(w, "Session expired", http.StatusGone)
		return nil, false
	}

	return &session, true
}
