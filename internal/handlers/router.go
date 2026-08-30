package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"slate-api/internal/auth"
	"slate-api/internal/config"
)

type RouterDeps struct {
	Cfg           *config.Config
	TokenService  *auth.TokenService
	AuthHandler   *AuthHandler
	GoogleHandler *GoogleHandler
	TitleHandler  *TitleHandler
}

func NewRouter(deps RouterDeps) chi.Router {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{deps.Cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Token-Delivery"},
		AllowCredentials: true,
	}))

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", deps.AuthHandler.Register)
		r.Post("/login", deps.AuthHandler.Login)
		r.Post("/refresh", deps.AuthHandler.Refresh)
		r.Post("/logout", deps.AuthHandler.Logout)
		r.Post("/google", deps.GoogleHandler.GoogleLogin)
	})

	r.Route("/titles", func(r chi.Router) {
		// public
		r.Get("/trending/movies", deps.TitleHandler.TrendingMovies)
		r.Get("/trending/tv", deps.TitleHandler.TrendingTV)

		// private
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireAuth(deps.TokenService))
			r.Get("/search", deps.TitleHandler.Search)
			r.Get("/", deps.TitleHandler.GetMyTitles)
			r.Post("/", deps.TitleHandler.RateTitle)
			r.Get("/{userTitleID}", deps.TitleHandler.GetDetail)
			r.Patch("/{userTitleID}", deps.TitleHandler.UpdateScore)
			r.Delete("/{userTitleID}", deps.TitleHandler.DeleteRating)
		})
	})

	return r
}