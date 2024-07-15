package main

import (
	"errors"
	"net"
	"net/http"

	mdw "sypchal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	config, err := GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("get config")
	}

	r := chi.NewRouter()

	// enable structured logging on prod
	if config.Environment == "production" {
		r.Use(mdw.Logger(log.Logger))
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Hostname, config.Port),
		Handler: r,
	}

	log.Printf("http server listening on %s", httpServer.Addr)
	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("serving http server")
	}
}
