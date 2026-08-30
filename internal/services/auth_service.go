package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"slate-api/internal/auth"
	"slate-api/internal/models"
)

type AuthService struct {
	db           *gorm.DB
	tokenService *auth.TokenService
}

func NewAuthService(db *gorm.DB, tokenService *auth.TokenService) *AuthService {
	return &AuthService{db: db, tokenService: tokenService}
}

var ErrEmailAlreadyRegistered = fmt.Errorf("email already registered")
var ErrInvalidCredentials = fmt.Errorf("invalid credentials")

func (s *AuthService) IssueTokens(user *models.User) (accessToken, rawRefresh string, expiresAt time.Time, err error) {
	accessToken, err = s.tokenService.CreateAccessToken(user)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("creating access token: %w", err)
	}

	rawRefresh, err = s.tokenService.GenerateRefreshToken()
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("generating refresh token: %w", err)
	}

	expiresAt = s.tokenService.RefreshTokenExpiresAt()

	rt := models.RefreshToken{
		UserID:    user.ID,
		TokenHash: s.tokenService.HashToken(rawRefresh),
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.db.Create(&rt).Error; err != nil {
		return "", "", time.Time{}, fmt.Errorf("saving refresh token: %w", err)
	}

	return accessToken, rawRefresh, expiresAt, nil
}

func (s *AuthService) Register(email, password string, fullName *string) (*models.User, error) {
	var existing models.User
	err := s.db.Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil, ErrEmailAlreadyRegistered
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("checking existing user: %w", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	user := models.User{
		Email:        email,
		PasswordHash: &hash,
		FullName:     fullName,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return &user, nil
}

func (s *AuthService) Login(email, password string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil || !auth.CheckPassword(*user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

func (s *AuthService) Refresh(incomingToken string) (*models.User, error) {
	tokenHash := s.tokenService.HashToken(incomingToken)

	var stored models.RefreshToken
	err := s.db.Where("token_hash = ?", tokenHash).First(&stored).Error
	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	if stored.RevokedAt != nil {
		now := time.Now().UTC()
		s.db.Model(&models.RefreshToken{}).
			Where("user_id = ? AND revoked_at IS NULL", stored.UserID).
			Update("revoked_at", now)
		return nil, auth.ErrInvalidToken
	}

	if stored.ExpiresAt.Before(time.Now().UTC()) {
		return nil, auth.ErrInvalidToken
	}

	now := time.Now().UTC()
	result := s.db.Model(&models.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", stored.ID).
		Update("revoked_at", now)
	if result.RowsAffected == 0 {
		return nil, auth.ErrInvalidToken
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", stored.UserID).Error; err != nil {
		return nil, auth.ErrInvalidToken
	}

	return &user, nil
}

func (s *AuthService) Logout(incomingToken string) error {
	if incomingToken == "" {
		return nil
	}
	tokenHash := s.tokenService.HashToken(incomingToken)

	now := time.Now().UTC()
	s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", now)

	return nil
}

func (s *AuthService) GetOrCreateGoogleUser(email, name, googleSubject string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		user = models.User{
			Email:     email,
			FullName:  &name,
			GoogleID:  &googleSubject,
			CreatedAt: time.Now().UTC(),
		}
		if err := s.db.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("creating google user: %w", err)
		}
		return &user, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user: %w", err)
	}

	if user.GoogleID == nil {
		user.GoogleID = &googleSubject
		s.db.Save(&user)
	}

	return &user, nil
}
