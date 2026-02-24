package handlers

import (
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maxbrt/steamzip-app/internal/utils"
)

func (h *Handler) HandleGetMasterImage(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}

	if session.Asset == nil || session.Asset.MasterFileKey == "" {
		http.Error(w, "No master image for this session", http.StatusNotFound)
		return
	}

	presigned, err := h.S3PresignClient.PresignGetObject(r.Context(), &s3.GetObjectInput{
		Bucket: aws.String(h.S3Bucket),
		Key:    aws.String(session.Asset.MasterFileKey),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		http.Error(w, "Error generating image URL", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"url": presigned.URL})
}
