package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserTitle struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index:idx_user_titles_user_id_title_id,unique"`
	TitleID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_titles_user_id_title_id,unique"`

	StoryScore     float64 `gorm:"default:0"`
	ActingScore    float64 `gorm:"default:0"`
	DirectionScore float64 `gorm:"default:0"`
	TechnicalScore float64 `gorm:"default:0"`
	ImpactScore    float64 `gorm:"default:0"`
	OverallScore   float64 `gorm:"default:0"`

	Review      *string `gorm:""`
	Recommended bool    `gorm:"default:false"`
	Completed   bool    `gorm:"default:true"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`

	Title Title `gorm:"foreignKey:TitleID"`
}

func (ut *UserTitle) BeforeCreate(tx *gorm.DB) error {
	if ut.ID == uuid.Nil {
		ut.ID = uuid.New()
	}
	return nil
}