package main

import (
	"context"
	"errors"
	"net/http"

	"sypchal/postgres"
	"sypchal/server"
	"sypchal/user"
	"sypchal/validation"

	"github.com/rs/zerolog/log"
)

func main() {
	config, err := GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("get config")
	}

	ctx := context.Background()

	validator := validation.NewValidator()

	db, err := postgres.NewPostgresClient(ctx, config.DatabaseUrl)
	if err != nil {
		log.Error().Err(err).Msg("new postgres client")
	}

	userDomain, err := user.NewUserDomain(db.Conn, validator, config.JwtSecret)
	if err != nil {
		log.Error().Err(err).Msg("new user domain")
	}

	httpServer, err := server.NewServer(server.ServerConfig{
		Environment: config.Environment,
		Hostname:    config.Hostname,
		Port:        config.Port,
		UserDomain:  userDomain,
	})
	if err != nil {
		log.Error().Err(err).Msg("new server")
	}

	log.Printf("http server listening on %s", httpServer.Addr)
	err = httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error().Err(err).Msg("serving http server")
	}
}
