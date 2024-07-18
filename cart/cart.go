package cart

import (
	"context"
	"errors"
	"sypchal/validation"
	"time"

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

type CartItem struct {
	Id        int        `json:"id"`
	UserId    int        `json:"user_id"`
	ProductId int        `json:"product_id"`
	Qty       int        `json:"qty"`
	Price     int        `json:"price"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
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

type Cart struct {
	TotalPrice    int                  `json:"total_price"`
	ItemCount     int                  `json:"item_count"`
	TotalQuantity int                  `json:"total_quantity"`
	Items         []*CartItemPopulated `json:"items"`
}

type CartItemPopulated struct {
	Id         int             `json:"id"`
	Product    CartItemProduct `json:"product"`
	Qty        int             `json:"qty"`
	Price      int             `json:"price"`
	TotalPrice int             `json:"total_price"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  *time.Time      `json:"updated_at"`
}

type CartItemProduct struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageUrl    string `json:"image_url"`
	Price       int    `json:"price"`
}

func (c *CartDomain) GetUserCart(ctx context.Context, userId int) (cart *Cart, err error) {
	rows, err := c.db.Query(
		ctx,
		`select 
			(cart_items.qty*cart_items.price) as total_price,
			products.id,
			products.name,
			products.description,
			products.image_url,
			products.price,
			cart_items.id,
			qty,
			cart_items.price,
			cart_items.created_at,
			cart_items.updated_at
		from cart_items inner join products on(product_id=products.id and user_id=$1);`,
		userId,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	cart = &Cart{}

	for rows.Next() {
		item := &CartItemPopulated{}
		rows.Scan(
			&item.TotalPrice,
			&item.Product.Id,
			&item.Product.Name,
			&item.Product.Description,
			&item.Product.ImageUrl,
			&item.Product.Price,
			&item.Id,
			&item.Qty,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		cart.ItemCount++
		cart.TotalPrice += item.TotalPrice
		cart.TotalQuantity += item.Qty
		cart.Items = append(cart.Items, item)
	}

	return
}
