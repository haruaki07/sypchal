package cart

import (
	"context"
	"errors"
	"sypchal/validation"

	"github.com/jackc/pgx/v5"
)

type CartDomain struct {
	db        *pgx.Conn
	validator *validation.Validator
}

func NewCartDomain(db *pgx.Conn, validator *validation.Validator) (*CartDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if validator == nil {
		return nil, errors.New("validator is nil")
	}

	return &CartDomain{db, validator}, nil
}

type AddCartItemRequest struct {
	UserId    int `json:"user_id" validate:"required"`
	ProductId int `json:"product_id" validate:"required"`
	Qty       int `json:"qty" validate:"required"`
	Price     int `json:"price" validate:"required"`
}

func (c *CartDomain) AddCartItem(ctx context.Context, req AddCartItemRequest) (count int, err error) {
	if err = c.validator.ValidateStruct(req); err != nil {
		return
	}

	// do upsert
	_, err = c.db.Exec(
		ctx,
		`insert into cart_items(user_id,product_id,qty,price) values ($1,$2,$3,$4)
		on conflict (user_id,product_id) do update set qty=excluded.qty+cart_items.qty`,
		req.UserId,
		req.ProductId,
		req.Qty,
		req.Price,
	)
	if err != nil {
		return
	}

	err = c.db.QueryRow(ctx, "select sum(qty) from cart_items where user_id=$1", req.UserId).Scan(&count)

	return
}
