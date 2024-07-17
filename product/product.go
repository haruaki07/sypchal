package product

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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
	Id          int        `json:"id"`
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

type UpdateProductRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ImageUrl    string `json:"image_url,omitempty"`
	Category    string `json:"category,omitempty"`
	Stock       int    `json:"stock,omitempty"`
	Price       int    `json:"price,omitempty"`
}

func (p *ProductDomain) UpdateProductById(ctx context.Context, id int, req UpdateProductRequest) (product *Product, err error) {
	if err = p.validator.ValidateStruct(req); err != nil {
		return
	}

	exists := p.IsProductExists(ctx, id)
	if !exists {
		err = ErrProductNotFound
		return
	}

	product = &Product{}
	fields := make([]string, 0, 6)
	args := make([]interface{}, 0, 7) // +1 for id
	rv := reflect.ValueOf(req)

	if rv.NumField() < 1 {
		err = errors.New("please fill atleast one field")
		return
	}

	if req.Name != "" {
		fields = append(fields, "name=$"+strconv.Itoa(len(fields)+1))
		args = append(args, req.Name)
	}

	if req.Description != "" {
		fields = append(fields, "description=$"+strconv.Itoa(len(fields)+1))
		args = append(args, req.Description)
	}

	if req.ImageUrl != "" {
		fields = append(fields, "image_url=$"+strconv.Itoa(len(fields)+1))
		args = append(args, req.ImageUrl)
	}

	if req.Category != "" {
		fields = append(fields, "category=$"+strconv.Itoa(len(fields)+1))
		args = append(args, req.Category)
	}

	if req.Stock != 0 {
		fields = append(fields, "stock=$"+strconv.Itoa(len(fields)+1))
		args = append(args, req.Stock)
	}

	if req.Price != 0 {
		fields = append(fields, "price=$"+strconv.Itoa(len(fields)+1))
		args = append(args, req.Price)
	}

	args = append(args, id)
	err = p.db.QueryRow(
		ctx,
		fmt.Sprintf(
			`update products set %s where id = $%d
			returning id,name,description,image_url,category,stock,price,created_at,updated_at`,
			strings.Join(fields, ","),
			len(args),
		),
		args...,
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

func (p *ProductDomain) IsProductExists(ctx context.Context, id int) bool {
	var count int
	_ = p.db.QueryRow(ctx, "select count(*) from products where id = $1", id).Scan(&count)

	return count > 0
}
