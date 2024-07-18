package server

import (
	"net"
	"net/http"
	"sypchal/cart"
	"sypchal/product"
	"sypchal/user"

	mdw "sypchal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
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
	CartDomain    *cart.CartDomain
}

type ServerDependency struct {
	userDomain    *user.UserDomain
	productDomain *product.ProductDomain
	cartDomain    *cart.CartDomain
}

func NewServer(config ServerConfig) (*http.Server, error) {
	dependencies := &ServerDependency{
		userDomain:    config.UserDomain,
		productDomain: config.ProductDomain,
		cartDomain:    config.CartDomain,
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
		r.Use(jwtauth.Verifier(dependencies.userDomain.Jwt))
		r.Use(jwtauth.Authenticator(dependencies.userDomain.Jwt))

		r.Get("/api/products", dependencies.ProductList)
		r.Get("/api/products/{id:^[0-9]*$}", dependencies.ProductGet)
		r.Get("/api/category/{category}", dependencies.ProductListByCategory)
		r.Get("/api/cart", dependencies.CartGet)
		r.Post("/api/cart", dependencies.CartAdd)
		r.Delete("/api/cart/{id:^[0-9]*$}", dependencies.CartDeleteItem)
		r.Put("/api/cart/{id:^[0-9]*$}", dependencies.CartUpdateItem)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.BasicAuth("admin area", map[string]string{
			config.Admin.Username: config.Admin.Password,
		}))

		r.Post("/api/products", dependencies.ProductCreate)
		r.Put("/api/products/{id:^[0-9]*$}", dependencies.ProductUpdate)
		r.Delete("/api/products/{id:^[0-9]*$}", dependencies.ProductDelete)
	})

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Hostname, config.Port),
		Handler: r,
	}

	return httpServer, nil
}
