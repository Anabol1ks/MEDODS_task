package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;index"`
	RefreshTokenHash string    `gorm:"not null"`
	UserAgent        string    `gorm:"not null"`
	IP               string    `gorm:"not null"`
	CreatedAt        time.Time
}
