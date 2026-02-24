package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetStatus = string

const (
	AssetStatusUploading  AssetStatus = "uploading"
	AssetStatusUploaded   AssetStatus = "uploaded"
	AssetStatusProcessing AssetStatus = "processing"
	AssetStatusReady      AssetStatus = "ready"
	AssetStatusFailed     AssetStatus = "failed"
)

type Asset struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;"`
	SessionID      uuid.UUID `gorm:"type:uuid;index"`
	MasterFileKey  string
	ZipKey         *string
	MimeType       string
	Status         AssetStatus
	MasterFileSize int64
	FocalPoints    *string   `gorm:"type:text"` // JSON-encoded focal points
}

func (a *Asset) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New()
	return
}
