package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash *string   `gorm:""`
	FullName     *string   `gorm:""`
	GoogleID     *string   `gorm:""`
	CreatedAt    time.Time `gorm:"not null"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}