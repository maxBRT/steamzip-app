package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/maxbrt/steamzip-app/internal/models"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/webhook"
	"gorm.io/gorm"
)

func (h *Handler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	// Read raw body before any parsing — required for signature verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), h.StripeWebhookSecret)
	if err != nil {
		log.Printf("stripe webhook signature verification failed: %v", err)
		http.Error(w, "Invalid signature", http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var cs stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &cs); err != nil {
			log.Printf("failed to unmarshal checkout.session.completed: %v", err)
			http.Error(w, "Failed to parse event", http.StatusBadRequest)
			return
		}

		sessionIDStr, ok := cs.Metadata["session_id"]
		if !ok {
			log.Printf("checkout.session.completed missing session_id metadata")
			w.WriteHeader(http.StatusOK)
			return
		}

		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			log.Printf("invalid session_id in stripe metadata: %s", sessionIDStr)
			w.WriteHeader(http.StatusOK)
			return
		}

		session := models.Session{}
		if err := h.Database.First(&session, sessionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("session not found for stripe webhook: %s", sessionIDStr)
			} else {
				log.Printf("error loading session %s: %v", sessionIDStr, err)
			}
			// Return 200 so Stripe doesn't retry indefinitely for bad data
			w.WriteHeader(http.StatusOK)
			return
		}

		// Mark session as paid (idempotent)
		if err := h.Database.Model(&session).Update("status", models.SessionStatusPaid).Error; err != nil {
			log.Printf("failed to mark session %s as paid: %v", sessionIDStr, err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Load the asset with a direct query
		var asset models.Asset
		if err := h.Database.Where("session_id = ?", sessionID).Order("created_at DESC").First(&asset).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("session %s has no asset; skipping processing", sessionIDStr)
			} else {
				log.Printf("error loading asset for session %s: %v", sessionIDStr, err)
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		session.Asset = &asset

		// Skip if already successfully processed
		if asset.Status == models.AssetStatusReady {
			log.Printf("asset %s already processed; skipping", asset.ID)
			w.WriteHeader(http.StatusOK)
			return
		}

		if session.Asset.FocalPoints == nil {
			log.Printf("asset %s has no focal points stored; marking failed", session.Asset.ID)
			if err := h.Database.Model(session.Asset).Update("status", models.AssetStatusFailed).Error; err != nil {
				log.Printf("failed to mark asset %s as failed: %v", session.Asset.ID, err)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Decode stored focal points
		var fp SteamAssetFocalPoints
		if err := json.Unmarshal([]byte(*session.Asset.FocalPoints), &fp); err != nil {
			log.Printf("failed to decode focal points for asset %s: %v", session.Asset.ID, err)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Mark asset as processing
		if err := h.Database.Model(session.Asset).Update("status", models.AssetStatusProcessing).Error; err != nil {
			log.Printf("failed to set asset %s to processing: %v", session.Asset.ID, err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Run synchronously so Stripe retries on failure (non-200)
		h.runProcessingJob(*session.Asset, fp.ToMap())

		// Verify it succeeded
		if err := h.Database.First(&asset, asset.ID).Error; err == nil && asset.Status != models.AssetStatusReady {
			log.Printf("processing job did not complete for asset %s", asset.ID)
			http.Error(w, "Processing failed", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
