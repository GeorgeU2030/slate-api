package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	CreatedAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time `gorm:""`
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}