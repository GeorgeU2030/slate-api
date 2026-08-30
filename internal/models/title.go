package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type TitleType string

const (
	TitleTypeMovie TitleType = "movie"
	TitleTypeTV    TitleType = "tv"
)

type Title struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey"`
	TmdbID              int            `gorm:"not null;index:idx_titles_tmdb_id_type,unique"`
	Type                TitleType      `gorm:"not null;index:idx_titles_tmdb_id_type,unique"`
	Name                string         `gorm:"size:500;not null"`
	OriginalTitle       *string        `gorm:"size:500"`
	Overview            *string        `gorm:""`
	PosterPath          *string        `gorm:"size:300"`
	BackdropPath        *string        `gorm:"size:300"`
	Date                *time.Time     `gorm:""`
	IsOngoing           bool           `gorm:"default:false"`
	SeasonsCount        *int           `gorm:""`
	Genres              pq.StringArray `gorm:"type:text[]"`
	ProductionCompanies pq.StringArray `gorm:"type:text[]"`
	WatchProviders      pq.StringArray `gorm:"type:text[]"`
	AverageScore        *float64       `gorm:""`
	RatingsCount        int            `gorm:"default:0"`
	CreatedAt           time.Time      `gorm:"not null"`
}

func (t *Title) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}