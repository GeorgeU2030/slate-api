package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"slate-api/internal/models"
)

type TitleBase struct {
	TmdbID              int              `json:"tmdbId"`
	Type                models.TitleType `json:"type"`
	Name                string           `json:"name"`
	OriginalTitle       *string          `json:"originalTitle"`
	Overview            *string          `json:"overview"`
	PosterURL           *string          `json:"posterUrl"`
	BackdropURL         *string          `json:"backdropUrl"`
	Date                *time.Time       `json:"date"`
	IsOngoing           bool             `json:"isOngoing"`
	SeasonsCount        *int             `json:"seasonsCount"`
	Genres              []string         `json:"genres"`
	ProductionCompanies []string         `json:"productionCompanies"`
	WatchProviders      []string         `json:"watchProviders"`
}

type TmdbSearchResult struct {
	TitleBase
	UserTitleID  *uuid.UUID `json:"userTitleId"`
	OverallScore *float64   `json:"overallScore"`
	Recommended  bool       `json:"recommended"`
	AverageScore *float64   `json:"averageScore"`
	RatingsCount int        `json:"ratingsCount"`
}

type TitleOut struct {
	TitleBase
	ID           uuid.UUID `json:"id"`
	UserTitleID  uuid.UUID `json:"userTitleId"`
	OverallScore float64   `json:"overallScore"`
	AverageScore *float64  `json:"averageScore"`
	RatingsCount int       `json:"ratingsCount"`
	Recommended  bool      `json:"recommended"`
	Completed    bool      `json:"completed"`
}

type TitleDetailOut struct {
	TitleBase
	ID             uuid.UUID `json:"id"`
	UserTitleID    uuid.UUID `json:"userTitleId"`
	StoryScore     float64   `json:"storyScore"`
	ActingScore    float64   `json:"actingScore"`
	DirectionScore float64   `json:"directionScore"`
	TechnicalScore float64   `json:"technicalScore"`
	ImpactScore    float64   `json:"impactScore"`
	OverallScore   float64   `json:"overallScore"`
	Review         *string   `json:"review"`
	Recommended    bool      `json:"recommended"`
	Completed      bool      `json:"completed"`
	AverageScore   *float64  `json:"averageScore"`
	RatingsCount   int       `json:"ratingsCount"`
}

type RateTitleRequest struct {
	TmdbID         int              `json:"tmdbId"`
	Type           models.TitleType `json:"type"`
	StoryScore     float64          `json:"storyScore"`
	ActingScore    float64          `json:"actingScore"`
	DirectionScore float64          `json:"directionScore"`
	TechnicalScore float64          `json:"technicalScore"`
	ImpactScore    float64          `json:"impactScore"`
	Review         *string          `json:"review"`
	Recommended    bool             `json:"recommended"`
	Completed      *bool            `json:"completed"`
}

func (r RateTitleRequest) Validate() error {
	return validateScores(r.StoryScore, r.ActingScore, r.DirectionScore, r.TechnicalScore, r.ImpactScore)
}

type UpdateScoreRequest struct {
	StoryScore     float64 `json:"storyScore"`
	ActingScore    float64 `json:"actingScore"`
	DirectionScore float64 `json:"directionScore"`
	TechnicalScore float64 `json:"technicalScore"`
	ImpactScore    float64 `json:"impactScore"`
	Review         *string `json:"review"`
	Recommended    bool    `json:"recommended"`
	Completed      *bool   `json:"completed"`
}

func (r UpdateScoreRequest) Validate() error {
	return validateScores(r.StoryScore, r.ActingScore, r.DirectionScore, r.TechnicalScore, r.ImpactScore)
}

func validateScores(scores ...float64) error {
	for _, v := range scores {
		if v < 0 || v > 10 {
			return fmt.Errorf("scores must be between 0 and 10")
		}
	}
	return nil
}