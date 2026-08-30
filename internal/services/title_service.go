package services

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"slate-api/internal/dto"
	"slate-api/internal/models"
	"slate-api/internal/scoring"
	"slate-api/internal/tmdb"
)

var ErrTitleNotFound = errors.New("rating not found")

type TitleService struct {
	db   *gorm.DB
	tmdb *tmdb.Service
}

func NewTitleService(db *gorm.DB, tmdbService *tmdb.Service) *TitleService {
	return &TitleService{db: db, tmdb: tmdbService}
}

// ---------- Trending ----------

func (s *TitleService) GetTrending(ctx context.Context, titleType models.TitleType) ([]dto.TmdbSearchResult, error) {
	var raw []tmdb.RawItem
	var err error
	if titleType == models.TitleTypeMovie {
		raw, err = s.tmdb.GetTrendingMovies(ctx)
	} else {
		raw, err = s.tmdb.GetTrendingTV(ctx)
	}
	if err != nil {
		return nil, err
	}
	if len(raw) > 10 {
		raw = raw[:10]
	}

	infos := s.tmdb.Enrich(ctx, raw, titleType == models.TitleTypeMovie)

	out := make([]dto.TmdbSearchResult, 0, len(raw))
	for i, item := range raw {
		info := infos[i]
		if info == nil {
			continue
		}
		out = append(out, toTmdbResult(info, item.ID, titleType, nil, nil, false, nil, 0))
	}
	return out, nil
}

// ---------- Search ----------

type ratingKey struct {
	TmdbID int
	Type   models.TitleType
}

type candidate struct {
	TmdbID int
	Type   models.TitleType
	Info   *tmdb.TitleInfo
}

func (s *TitleService) Search(ctx context.Context, userID *uuid.UUID, query string) ([]dto.TmdbSearchResult, error) {
	results, err := s.tmdb.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(results) > 10 {
		results = results[:10]
	}

	candidates := s.resolveCandidates(ctx, results)
	if len(candidates) == 0 {
		return []dto.TmdbSearchResult{}, nil
	}

	tmdbIDs := make([]int, 0, len(candidates))
	for _, c := range candidates {
		tmdbIDs = append(tmdbIDs, c.TmdbID)
	}

	communityStats, err := s.getCommunityStats(tmdbIDs)
	if err != nil {
		return nil, err
	}

	var myRatings map[ratingKey]myRating
	if userID != nil {
		myRatings, err = s.getMyRatings(*userID, tmdbIDs)
		if err != nil {
			return nil, err
		}
	}

	out := make([]dto.TmdbSearchResult, 0, len(candidates))
	for _, c := range candidates {
		key := ratingKey{TmdbID: c.TmdbID, Type: c.Type}

		var userTitleID *uuid.UUID
		var overallScore *float64
		recommended := false
		if mine, ok := myRatings[key]; ok {
			id := mine.ID
			score := mine.OverallScore
			userTitleID, overallScore, recommended = &id, &score, mine.Recommended
		}

		var averageScore *float64
		ratingsCount := 0
		if community, ok := communityStats[key]; ok {
			averageScore = community.AverageScore
			ratingsCount = community.RatingsCount
		}

		out = append(out, toTmdbResult(c.Info, c.TmdbID, c.Type, userTitleID, overallScore, recommended, averageScore, ratingsCount))
	}
	return out, nil
}

func (s *TitleService) resolveCandidates(ctx context.Context, items []tmdb.RawItem) []candidate {
	results := make([]*candidate, len(items))
	var wg sync.WaitGroup

	for i, item := range items {
		var titleType models.TitleType
		switch item.MediaType {
		case "movie":
			titleType = models.TitleTypeMovie
		case "tv":
			titleType = models.TitleTypeTV
		default:
			continue 
		}

		wg.Add(1)
		go func(i int, item tmdb.RawItem, titleType models.TitleType) {
			defer wg.Done()
			var info *tmdb.TitleInfo
			var err error
			if titleType == models.TitleTypeMovie {
				info, err = s.tmdb.GetMovieDetail(ctx, item.ID)
			} else {
				info, err = s.tmdb.GetTVDetail(ctx, item.ID)
			}
			if err != nil {
				return
			}
			results[i] = &candidate{TmdbID: item.ID, Type: titleType, Info: info}
		}(i, item, titleType)
	}

	wg.Wait()

	out := make([]candidate, 0, len(items))
	for _, c := range results {
		if c != nil {
			out = append(out, *c)
		}
	}
	return out
}

type communityStat struct {
	AverageScore *float64
	RatingsCount int
}

func (s *TitleService) getCommunityStats(tmdbIDs []int) (map[ratingKey]communityStat, error) {
	var rows []models.Title
	if err := s.db.Where("tmdb_id IN ?", tmdbIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[ratingKey]communityStat, len(rows))
	for _, row := range rows {
		out[ratingKey{TmdbID: row.TmdbID, Type: row.Type}] = communityStat{
			AverageScore: row.AverageScore,
			RatingsCount: row.RatingsCount,
		}
	}
	return out, nil
}

type myRating struct {
	ID           uuid.UUID
	OverallScore float64
	Recommended  bool
}

func (s *TitleService) getMyRatings(userID uuid.UUID, tmdbIDs []int) (map[ratingKey]myRating, error) {
	type row struct {
		ID           uuid.UUID
		TmdbID       int
		Type         models.TitleType
		OverallScore float64
		Recommended  bool
	}
	var rows []row
	err := s.db.Table("user_titles").
		Select("user_titles.id, titles.tmdb_id, titles.type, user_titles.overall_score, user_titles.recommended").
		Joins("JOIN titles ON titles.id = user_titles.title_id").
		Where("user_titles.user_id = ? AND titles.tmdb_id IN ?", userID, tmdbIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[ratingKey]myRating, len(rows))
	for _, r := range rows {
		out[ratingKey{TmdbID: r.TmdbID, Type: r.Type}] = myRating{ID: r.ID, OverallScore: r.OverallScore, Recommended: r.Recommended}
	}
	return out, nil
}

// ---------- My titles ----------

func (s *TitleService) GetMyTitles(userID uuid.UUID, year *int, titleType *models.TitleType, search *string) ([]dto.TitleOut, error) {
	var rows []models.UserTitle
	query := s.db.Preload("Title").Where("user_titles.user_id = ?", userID)

	needsJoin := year != nil || titleType != nil || (search != nil && strings.TrimSpace(*search) != "")
	if needsJoin {
		query = query.Joins("JOIN titles ON titles.id = user_titles.title_id")
	}
	if year != nil {
		query = query.Where("EXTRACT(YEAR FROM titles.date) = ?", *year)
	}
	if titleType != nil {
		query = query.Where("titles.type = ?", *titleType)
	}
	if search != nil && strings.TrimSpace(*search) != "" {
		query = query.Where("titles.name ILIKE ?", "%"+*search+"%")
	}

	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]dto.TitleOut, 0, len(rows))
	for _, ut := range rows {
		out = append(out, toTitleOut(&ut.Title, &ut))
	}
	return out, nil
}

// ---------- Rate ----------

func (s *TitleService) Rate(ctx context.Context, userID uuid.UUID, body dto.RateTitleRequest) (*dto.TitleOut, error) {
	var title models.Title
	err := s.db.Where("tmdb_id = ? AND type = ?", body.TmdbID, body.Type).First(&title).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		var info *tmdb.TitleInfo
		if body.Type == models.TitleTypeMovie {
			info, err = s.tmdb.GetMovieDetail(ctx, body.TmdbID)
		} else {
			info, err = s.tmdb.GetTVDetail(ctx, body.TmdbID)
		}
		if err != nil {
			return nil, err
		}

		title = toNewTitle(info, body.TmdbID, body.Type)
		if err := s.db.Create(&title).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	var ut models.UserTitle
	err = s.db.Where("user_id = ? AND title_id = ?", userID, title.ID).First(&ut).Error
	isNew := errors.Is(err, gorm.ErrRecordNotFound)
	if isNew {
		ut = models.UserTitle{UserID: userID, TitleID: title.ID}
	} else if err != nil {
		return nil, err
	}

	applyScores(&ut, body.StoryScore, body.ActingScore, body.DirectionScore, body.TechnicalScore, body.ImpactScore, body.Review, body.Recommended, body.Completed)

	now := time.Now().UTC()
	ut.UpdatedAt = now
	if isNew {
		ut.CreatedAt = now
		if err := s.db.Create(&ut).Error; err != nil {
			return nil, err
		}
	} else if err := s.db.Save(&ut).Error; err != nil {
		return nil, err
	}

	if err := s.updateTitleAggregate(&title, &ut.ID, &ut.OverallScore); err != nil {
		return nil, err
	}

	result := toTitleOut(&title, &ut)
	return &result, nil
}

func applyScores(ut *models.UserTitle, story, acting, direction, technical, impact float64, review *string, recommended bool, completed *bool) {
	ut.StoryScore = story
	ut.ActingScore = acting
	ut.DirectionScore = direction
	ut.TechnicalScore = technical
	ut.ImpactScore = impact
	ut.Review = review
	ut.Recommended = recommended
	if completed != nil {
		ut.Completed = *completed
	} else {
		ut.Completed = true
	}
	ut.OverallScore = scoring.Calculate(story, acting, direction, technical, impact)
}

func (s *TitleService) updateTitleAggregate(title *models.Title, excludeUserTitleID *uuid.UUID, includeScore *float64) error {
	query := s.db.Model(&models.UserTitle{}).Where("title_id = ?", title.ID)
	if excludeUserTitleID != nil {
		query = query.Where("id != ?", *excludeUserTitleID)
	}

	var scores []float64
	if err := query.Pluck("overall_score", &scores).Error; err != nil {
		return err
	}
	if includeScore != nil {
		scores = append(scores, *includeScore)
	}

	title.RatingsCount = len(scores)
	if len(scores) == 0 {
		title.AverageScore = nil
	} else {
		sum := 0.0
		for _, sc := range scores {
			sum += sc
		}
		avg := sum / float64(len(scores))
		title.AverageScore = &avg
	}

	return s.db.Save(title).Error
}

// ---------- Detail / Update / Delete ----------

func (s *TitleService) GetDetail(userID, userTitleID uuid.UUID) (*dto.TitleDetailOut, error) {
	ut, err := s.getOwnedUserTitle(userID, userTitleID)
	if err != nil {
		return nil, err
	}
	result := toTitleDetailOut(&ut.Title, ut)
	return &result, nil
}

func (s *TitleService) UpdateScore(userID, userTitleID uuid.UUID, body dto.UpdateScoreRequest) (*dto.TitleOut, error) {
	ut, err := s.getOwnedUserTitle(userID, userTitleID)
	if err != nil {
		return nil, err
	}

	applyScores(ut, body.StoryScore, body.ActingScore, body.DirectionScore, body.TechnicalScore, body.ImpactScore, body.Review, body.Recommended, body.Completed)
	ut.UpdatedAt = time.Now().UTC()

	if err := s.db.Save(ut).Error; err != nil {
		return nil, err
	}
	if err := s.updateTitleAggregate(&ut.Title, &ut.ID, &ut.OverallScore); err != nil {
		return nil, err
	}

	result := toTitleOut(&ut.Title, ut)
	return &result, nil
}

func (s *TitleService) DeleteRating(userID, userTitleID uuid.UUID) error {
	ut, err := s.getOwnedUserTitle(userID, userTitleID)
	if err != nil {
		return err
	}

	if err := s.db.Delete(ut).Error; err != nil {
		return err
	}
	return s.updateTitleAggregate(&ut.Title, nil, nil)
}

func (s *TitleService) getOwnedUserTitle(userID, userTitleID uuid.UUID) (*models.UserTitle, error) {
	var ut models.UserTitle
	err := s.db.Preload("Title").
		Where("id = ? AND user_id = ?", userTitleID, userID).
		First(&ut).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTitleNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ut, nil
}

// ---------- Mappers ----------

func toNewTitle(info *tmdb.TitleInfo, tmdbID int, titleType models.TitleType) models.Title {
	return models.Title{
		TmdbID:              tmdbID,
		Type:                titleType,
		Name:                info.Name,
		OriginalTitle:       info.OriginalTitle,
		Overview:            info.Overview,
		PosterPath:          info.PosterPath,
		BackdropPath:        info.BackdropPath,
		Date:                info.Date,
		IsOngoing:           info.IsOngoing,
		SeasonsCount:        info.SeasonsCount,
		Genres:              info.Genres,
		ProductionCompanies: info.ProductionCompanies,
		WatchProviders:      info.WatchProviders,
		CreatedAt:           time.Now().UTC(),
	}
}

func toTmdbResult(info *tmdb.TitleInfo, tmdbID int, titleType models.TitleType, userTitleID *uuid.UUID, overallScore *float64, recommended bool, averageScore *float64, ratingsCount int) dto.TmdbSearchResult {
	return dto.TmdbSearchResult{
		TitleBase: dto.TitleBase{
			TmdbID: tmdbID, Type: titleType, Name: info.Name,
			OriginalTitle: info.OriginalTitle, Overview: info.Overview,
			PosterURL: tmdb.PosterURL(info.PosterPath), BackdropURL: tmdb.BackdropURL(info.BackdropPath),
			Date: info.Date, IsOngoing: info.IsOngoing, SeasonsCount: info.SeasonsCount,
			Genres: info.Genres, ProductionCompanies: info.ProductionCompanies, WatchProviders: info.WatchProviders,
		},
		UserTitleID: userTitleID, OverallScore: overallScore, Recommended: recommended,
		AverageScore: averageScore, RatingsCount: ratingsCount,
	}
}

func toTitleOut(title *models.Title, ut *models.UserTitle) dto.TitleOut {
	return dto.TitleOut{
		TitleBase: dto.TitleBase{
			TmdbID: title.TmdbID, Type: title.Type, Name: title.Name,
			OriginalTitle: title.OriginalTitle, Overview: title.Overview,
			PosterURL: tmdb.PosterURL(title.PosterPath), BackdropURL: tmdb.BackdropURL(title.BackdropPath),
			Date: title.Date, IsOngoing: title.IsOngoing, SeasonsCount: title.SeasonsCount,
			Genres: []string(title.Genres), ProductionCompanies: []string(title.ProductionCompanies),
			WatchProviders: []string(title.WatchProviders),
		},
		ID: title.ID, UserTitleID: ut.ID, OverallScore: ut.OverallScore,
		AverageScore: title.AverageScore, RatingsCount: title.RatingsCount,
		Recommended: ut.Recommended, Completed: ut.Completed,
	}
}

func toTitleDetailOut(title *models.Title, ut *models.UserTitle) dto.TitleDetailOut {
	return dto.TitleDetailOut{
		TitleBase: dto.TitleBase{
			TmdbID: title.TmdbID, Type: title.Type, Name: title.Name,
			OriginalTitle: title.OriginalTitle, Overview: title.Overview,
			PosterURL: tmdb.PosterURL(title.PosterPath), BackdropURL: tmdb.BackdropURL(title.BackdropPath),
			Date: title.Date, IsOngoing: title.IsOngoing, SeasonsCount: title.SeasonsCount,
			Genres: []string(title.Genres), ProductionCompanies: []string(title.ProductionCompanies),
			WatchProviders: []string(title.WatchProviders),
		},
		ID: title.ID, UserTitleID: ut.ID,
		StoryScore: ut.StoryScore, ActingScore: ut.ActingScore, DirectionScore: ut.DirectionScore,
		TechnicalScore: ut.TechnicalScore, ImpactScore: ut.ImpactScore, OverallScore: ut.OverallScore,
		Review: ut.Review, Recommended: ut.Recommended, Completed: ut.Completed,
		AverageScore: title.AverageScore, RatingsCount: title.RatingsCount,
	}
}