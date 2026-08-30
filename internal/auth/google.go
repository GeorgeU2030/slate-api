package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var ErrGoogleAuth = errors.New("google auth failed")

type GooglePayload struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

type GoogleAuthService struct {
	oauthConfig *oauth2.Config
}

func NewGoogleAuthService(clientID, clientSecret string) *GoogleAuthService {
	return &GoogleAuthService{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     googleoauth.Endpoint,
			RedirectURL:  "postmessage", 
		},
	}
}

func (s *GoogleAuthService) ExchangeCodeForIDToken(ctx context.Context, code string) (string, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("%w: calling google token endpoint: %v", ErrGoogleAuth, err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", fmt.Errorf("%w: google response missing id_token", ErrGoogleAuth)
	}

	return rawIDToken, nil
}

func (s *GoogleAuthService) ValidateIDToken(ctx context.Context, rawIDToken, audience string) (*GooglePayload, error) {
	payload, err := idtoken.Validate(ctx, rawIDToken, audience)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid google token: %v", ErrGoogleAuth, err)
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)

	return &GooglePayload{
		Subject:       payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
	}, nil
}