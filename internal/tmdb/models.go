package tmdb

import "time"

type Genre struct {
	Name string `json:"name"`
}

type Company struct {
	Name string `json:"name"`
}

type Provider struct {
	ProviderName string `json:"provider_name"`
}

type WatchProviderRegion struct {
	Flatrate []Provider `json:"flatrate"`
}

type WatchProviders struct {
	Results map[string]WatchProviderRegion `json:"results"`
}

type MovieDetail struct {
	ID                  int            `json:"id"`
	Title               string         `json:"title"`
	OriginalTitle       string         `json:"original_title"`
	Overview            string         `json:"overview"`
	PosterPath          *string        `json:"poster_path"`
	BackdropPath        *string        `json:"backdrop_path"`
	ReleaseDate         Date           `json:"release_date"`
	Genres              []Genre        `json:"genres"`
	ProductionCompanies []Company      `json:"production_companies"`
	WatchProviders      WatchProviders `json:"watch/providers"`
}

type TvDetail struct {
	ID              int            `json:"id"`
	Name            string         `json:"name"`
	OriginalName    string         `json:"original_name"`
	Overview        string         `json:"overview"`
	PosterPath      *string        `json:"poster_path"`
	BackdropPath    *string        `json:"backdrop_path"`
	LastAirDate     Date           `json:"last_air_date"`
	Status          string         `json:"status"`
	NumberOfSeasons *int           `json:"number_of_seasons"`
	Genres          []Genre        `json:"genres"`
	Networks        []Company      `json:"networks"`
	WatchProviders  WatchProviders `json:"watch/providers"`
}

type RawItem struct {
	ID        int    `json:"id"`
	MediaType string `json:"media_type"`
}

type TitleInfo struct {
	Name                string
	OriginalTitle       *string
	Overview            *string
	PosterPath          *string
	BackdropPath        *string
	Date                *time.Time
	IsOngoing           bool
	SeasonsCount        *int
	Genres              []string
	ProductionCompanies []string
	WatchProviders      []string
}