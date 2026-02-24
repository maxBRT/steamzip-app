package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maxbrt/steamzip-app/internal/models"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

type GetUploadURLRequest struct {
	ContentType string `json:"contentType"`
}

type GetUploadURLResponse struct {
	AssetID   string `json:"assetId"`
	UploadURL string `json:"uploadUrl"`
}

func (h *Handler) HandleGetUploadURL(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}

	var body GetUploadURLRequest
	if err := utils.ReadJSON(r, &body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	allowed := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
	}
	if !allowed[body.ContentType] {
		http.Error(w, "unsupported content type", http.StatusBadRequest)
		return
	}

	// Reject if the session already has an asset in a non-terminal state.
	var existing models.Asset
	err := h.Database.Where("session_id = ? AND status NOT IN ?", session.ID,
		[]string{models.AssetStatusFailed}).First(&existing).Error
	if err == nil {
		http.Error(w, "Session already has an active asset", http.StatusConflict)
		return
	}

	asset := models.Asset{
		SessionID: session.ID,
		MimeType:  body.ContentType,
		Status:    models.AssetStatusUploading,
	}

	if err := h.Database.Create(&asset).Error; err != nil {
		log.Println(err)
		http.Error(w, "Error creating asset", http.StatusInternalServerError)
		return
	}

	s3Key := fmt.Sprintf("masters/%s", asset.ID)

	if err := h.Database.Model(&asset).Update("master_file_key", s3Key).Error; err != nil {
		log.Println(err)
		http.Error(w, "Error updating asset", http.StatusInternalServerError)
		return
	}

	presigned, err := h.S3PresignClient.PresignPutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(h.S3Bucket),
		Key:         aws.String(s3Key),
		ContentType: aws.String(body.ContentType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		log.Println(err)
		http.Error(w, "Error generating upload URL", http.StatusInternalServerError)
		return
	}

	if err := utils.WriteJSON(w, http.StatusOK, GetUploadURLResponse{
		AssetID:   asset.ID.String(),
		UploadURL: presigned.URL,
	}); err != nil {
		log.Println(err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
	}
}
