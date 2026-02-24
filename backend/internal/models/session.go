package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionStatus = string

const (
	SessionStatusPaymentPending SessionStatus = "payment_pending"
	SessionStatusPaid           SessionStatus = "paid"
)

type Session struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;"`
	Status      SessionStatus `gorm:"default:payment_pending"`
	DownloadURL *string
	Asset       *Asset    `gorm:"foreignKey:SessionID"`
	ExpiresAt time.Time
}

func (s *Session) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New()
	if s.ExpiresAt.IsZero() {
		s.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	return
}
