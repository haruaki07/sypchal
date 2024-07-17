package server

import (
	"net"
	"net/http"
	"sypchal/product"
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
	Admin       struct {
		Username string
		Password string
	}
	UserDomain    *user.UserDomain
	ProductDomain *product.ProductDomain
}

type ServerDependency struct {
	userDomain    *user.UserDomain
	productDomain *product.ProductDomain
}

func NewServer(config ServerConfig) (*http.Server, error) {
	dependencies := &ServerDependency{
		userDomain:    config.UserDomain,
		productDomain: config.ProductDomain,
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

	r.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("admin area", map[string]string{
			config.Admin.Username: config.Admin.Password,
		}))

		r.Post("/api/products", dependencies.ProductCreate)
		r.Put("/api/products/{id:[0-9]}", dependencies.ProductUpdate)
	})

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Hostname, config.Port),
		Handler: r,
	}

	return httpServer, nil
}
