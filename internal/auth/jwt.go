package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"slate-api/internal/models"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenService(secret string, accessMinutes, refreshDays int) *TokenService {
	return &TokenService{
		secret:     []byte(secret),
		accessTTL:  time.Duration(accessMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

type AccessClaims struct {
	Email    string `json:"email"`
	FullName string `json:"full_name,omitempty"`
	jwt.RegisteredClaims
}

func (s *TokenService) CreateAccessToken(user *models.User) (string, error) {
	now := time.Now()

	claims := AccessClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
			ID:        uuid.NewString(), 
		},
	}
	if user.FullName != nil {
		claims.FullName = *user.FullName
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *TokenService) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *TokenService) GenerateRefreshToken() (string, error) {
	return generateRandomURLSafeString(64)
}

func (s *TokenService) HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *TokenService) RefreshTokenExpiresAt() time.Time {
	return time.Now().UTC().Add(s.refreshTTL)
}