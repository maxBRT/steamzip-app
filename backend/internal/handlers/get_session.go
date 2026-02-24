package handlers

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

type GetSessionResponse struct {
	SessionID   string  `json:"sessionId"`
	Status      string  `json:"status"`
	AssetStatus *string `json:"assetStatus,omitempty"`
	DownloadUrl string  `json:"downloadUrl,omitempty"`
}

func (h *Handler) HandleGetSession(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}

	response := GetSessionResponse{
		SessionID: session.ID.String(),
		Status:    session.Status,
	}

	if session.Asset != nil {
		response.AssetStatus = &session.Asset.Status
	}

	// Only generate a download URL once processing is complete
	if session.Asset != nil && session.Asset.ZipKey != nil {
		presigned, err := h.S3PresignClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
			Bucket: aws.String(h.S3Bucket),
			Key:    aws.String(*session.Asset.ZipKey),
		}, s3.WithPresignExpires(time.Until(session.ExpiresAt)))
		if err != nil {
			http.Error(w, "Error generating download URL", http.StatusInternalServerError)
			return
		}
		response.DownloadUrl = presigned.URL
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
