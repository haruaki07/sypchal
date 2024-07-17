package server

import (
	"net"
	"net/http"
	"sypchal/user"

	mdw "sypchal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type ServerConfig struct {
	Hostname    string
	Port        string
	Environment string
	UserDomain  *user.UserDomain
}

type ServerDependency struct {
	userDomain *user.UserDomain
}

func NewServer(config ServerConfig) (*http.Server, error) {
	dependencies := &ServerDependency{
		userDomain: config.UserDomain,
	}

	r := chi.NewRouter()

	// enable structured logging on prod
	if config.Environment == "production" {
		r.Use(mdw.Logger(log.Logger))
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Post("/api/register", dependencies.UserRegister)
	r.Post("/api/login", dependencies.UserLogin)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Hostname, config.Port),
		Handler: r,
	}

	return httpServer, nil
}
