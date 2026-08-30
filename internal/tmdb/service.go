package tmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	posterBaseURL   = "https://image.tmdb.org/t/p/w500"
	backdropBaseURL = "https://image.tmdb.org/t/p/original"
)

var endedStatuses = map[string]bool{"Ended": true, "Canceled": true}

func PosterURL(path *string) *string {
	if path == nil || *path == "" {
		return nil
	}
	u := posterBaseURL + *path
	return &u
}

func BackdropURL(path *string) *string {
	if path == nil || *path == "" {
		return nil
	}
	u := backdropBaseURL + *path
	return &u
}

type Service struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewService(apiKey string) *Service {
	return &Service{
		apiKey:     apiKey,
		baseURL:    "https://api.themoviedb.org/3/",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *Service) getJSON(ctx context.Context, path string, dst any) error {
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	fullURL := fmt.Sprintf("%s%s%sapi_key=%s&language=en-US", s.baseURL, path, sep, url.QueryEscape(s.apiKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("building tmdb request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("calling tmdb: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tmdb returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decoding tmdb response: %w", err)
	}
	return nil
}

func (s *Service) fetchResults(ctx context.Context, path string) ([]RawItem, error) {
	var envelope struct {
		Results []RawItem `json:"results"`
	}
	if err := s.getJSON(ctx, path, &envelope); err != nil {
		return nil, err
	}
	return envelope.Results, nil
}

func (s *Service) GetTrendingMovies(ctx context.Context) ([]RawItem, error) {
	return s.fetchResults(ctx, "trending/movie/week")
}

func (s *Service) GetTrendingTV(ctx context.Context) ([]RawItem, error) {
	return s.fetchResults(ctx, "trending/tv/week")
}

func (s *Service) Search(ctx context.Context, query string) ([]RawItem, error) {
	return s.fetchResults(ctx, "search/multi?query="+url.QueryEscape(query))
}

func (s *Service) GetMovieDetail(ctx context.Context, tmdbID int) (*TitleInfo, error) {
	var detail MovieDetail
	path := fmt.Sprintf("movie/%d?append_to_response=watch/providers", tmdbID)
	if err := s.getJSON(ctx, path, &detail); err != nil {
		return nil, err
	}
	return mapMovie(&detail), nil
}

func (s *Service) GetTVDetail(ctx context.Context, tmdbID int) (*TitleInfo, error) {
	var detail TvDetail
	path := fmt.Sprintf("tv/%d?append_to_response=watch/providers", tmdbID)
	if err := s.getJSON(ctx, path, &detail); err != nil {
		return nil, err
	}
	return mapTv(&detail), nil
}

func mapMovie(d *MovieDetail) *TitleInfo {
	return &TitleInfo{
		Name:                d.Title,
		OriginalTitle:       nonEmptyPtr(d.OriginalTitle),
		Overview:            nonEmptyPtr(d.Overview),
		PosterPath:          d.PosterPath,
		BackdropPath:        d.BackdropPath,
		Date:                d.ReleaseDate.Ptr(),
		IsOngoing:           false,
		Genres:              genreNames(d.Genres),
		ProductionCompanies: companyNames(d.ProductionCompanies, 3),
		WatchProviders:      extractProviders(d.WatchProviders),
	}
}

func mapTv(d *TvDetail) *TitleInfo {
	isOngoing := d.Status != "" && !endedStatuses[d.Status]
	return &TitleInfo{
		Name:                d.Name,
		OriginalTitle:       nonEmptyPtr(d.OriginalName),
		Overview:            nonEmptyPtr(d.Overview),
		PosterPath:          d.PosterPath,
		BackdropPath:        d.BackdropPath,
		Date:                d.LastAirDate.Ptr(),
		IsOngoing:           isOngoing,
		SeasonsCount:        d.NumberOfSeasons,
		Genres:              genreNames(d.Genres),
		ProductionCompanies: companyNames(d.Networks, 3),
		WatchProviders:      extractProviders(d.WatchProviders),
	}
}

func genreNames(genres []Genre) []string {
	out := make([]string, 0, len(genres))
	for _, g := range genres {
		out = append(out, g.Name)
	}
	return out
}

func companyNames(companies []Company, limit int) []string {
	if len(companies) > limit {
		companies = companies[:limit]
	}
	out := make([]string, 0, len(companies))
	for _, c := range companies {
		out = append(out, c.Name)
	}
	return out
}

func extractProviders(wp WatchProviders) []string {
	region, ok := wp.Results["US"]
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(region.Flatrate))
	for _, p := range region.Flatrate {
		out = append(out, p.ProviderName)
	}
	return out
}

func nonEmptyPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *Service) Enrich(ctx context.Context, items []RawItem, isMovie bool) []*TitleInfo {
	results := make([]*TitleInfo, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		wg.Add(1)
		go func(i int, item RawItem) {
			defer wg.Done()
			var info *TitleInfo
			var err error
			if isMovie {
				info, err = s.GetMovieDetail(ctx, item.ID)
			} else {
				info, err = s.GetTVDetail(ctx, item.ID)
			}
			if err != nil {
				return 
			}
			results[i] = info
		}(i, item)
	}

	wg.Wait()
	return results
}

func (s *Service) Close() {
	s.httpClient.CloseIdleConnections()
}