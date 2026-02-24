package handlers

import (
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/maxbrt/steamzip-app/internal/models"
)

func (h *Handler) HandleConfirmUpload(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}

	if session.Asset == nil || session.Asset.Status != models.AssetStatusUploading {
		http.Error(w, "Asset is not in uploading state", http.StatusConflict)
		return
	}

	// Verify the file actually landed in S3 before accepting the transition
	obj, err := h.S3Client.HeadObject(r.Context(), &s3.HeadObjectInput{
		Bucket: aws.String(h.S3Bucket),
		Key:    aws.String(session.Asset.MasterFileKey),
	})
	if err != nil {
		http.Error(w, "File not found in storage — upload may not have completed", http.StatusConflict)
		return
	}

	const maxBytes = 50 * 1024 * 1024
	if obj.ContentLength != nil && *obj.ContentLength > maxBytes {
		http.Error(w, "file exceeds 50 MB limit", http.StatusRequestEntityTooLarge)
		return
	}

	if err := h.Database.Model(session.Asset).Update("status", models.AssetStatusUploaded).Error; err != nil {
		log.Println(err)
		http.Error(w, "Error updating asset status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
