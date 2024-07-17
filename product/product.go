package product

import (
	"context"
	"errors"
	"sypchal/validation"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProductDomain struct {
	db        *pgx.Conn
	validator *validation.Validator
}

func NewProductDomain(db *pgx.Conn, validator *validation.Validator) (*ProductDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &ProductDomain{db, validator}, nil
}

type Product struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ImageUrl    string     `json:"image_url"`
	Category    string     `json:"category"`
	Stock       int        `json:"stock"`
	Price       int        `json:"price"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	ImageUrl    string `json:"image_url"`
	Category    string `json:"category"`
	Stock       int    `json:"stock" validate:"required"`
	Price       int    `json:"price" validate:"required"`
}

func (p *ProductDomain) CreateProduct(ctx context.Context, req CreateProductRequest) (product *Product, err error) {
	if err = p.validator.ValidateStruct(req); err != nil {
		return
	}

	product = &Product{}
	err = p.db.QueryRow(
		ctx,
		`insert into products(name,description,image_url,category,stock,price) values ($1,$2,$3,$4,$5,$6) 
		returning id,name,description,image_url,category,stock,price,created_at,updated_at`,
		req.Name,
		req.Description,
		req.ImageUrl,
		req.Category,
		req.Stock,
		req.Price,
	).Scan(
		&product.Id,
		&product.Name,
		&product.Description,
		&product.ImageUrl,
		&product.Category,
		&product.Stock,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return
	}

	return
}
