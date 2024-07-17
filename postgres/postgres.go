package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type PostgresClient struct {
	Conn *pgx.Conn
}

func NewPostgresClient(ctx context.Context, dbUrl string) (*PostgresClient, error) {
	if dbUrl == "" {
		return nil, errors.New("dbUrl must be provided")
	}

	conn, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	return &PostgresClient{conn}, nil
}
