package order

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type OrderDomain struct {
	db *pgx.Conn
}

func NewOrderDomain(db *pgx.Conn) (*OrderDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &OrderDomain{db}, nil
}

var (
	OrderStatusUnpaid = "unpaid"
	OrderStatusPaid   = "paid"
)

type Order struct {
	Id         int        `json:"id"`
	UserId     int        `json:"user_id"`
	TotalPrice int        `json:"total_price"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

type CartItem struct {
	TotalPrice   int
	ProductId    int
	ProductStock int
	Qty          int
	Price        int
}

func (o *OrderDomain) PlaceOrder(ctx context.Context, userId int) (err error) {
	tx, err := o.db.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	// get cart items and products detail
	rows, err := tx.Query(
		ctx,
		`select 
			(cart_items.qty*cart_items.price) as total_price,
			products.id,
			products.stock,
			cart_items.qty,
			cart_items.price
		from cart_items inner join products on(product_id=products.id and user_id=$1);`,
		userId,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	var orderTotalPrice int
	items := []*CartItem{}
	for rows.Next() {
		item := &CartItem{}
		rows.Scan(
			&item.TotalPrice,
			&item.ProductId,
			&item.ProductStock,
			&item.Qty,
			&item.Price,
		)
		items = append(items, item)

		if item.Qty > item.ProductStock {
			err = ErrItemOutOfStock
			return
		}

		orderTotalPrice += item.TotalPrice
	}

	// create order entry
	payId := randStr(8)
	order := &Order{}
	err = tx.QueryRow(
		ctx,
		`insert into orders (user_id,total_price,status,pay_id) values ($1,$2,$3,$4) 
		returning id,user_id,total_price,status,created_at,updated_at`,
		userId,
		orderTotalPrice,
		OrderStatusUnpaid,
		payId,
	).Scan(
		&order.Id,
		&order.UserId,
		&order.TotalPrice,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return
	}

	// batch insert the order items
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"order_items"},
		[]string{"order_id", "product_id", "qty", "price"},
		pgx.CopyFromSlice(len(items), func(i int) ([]any, error) {
			return []any{
				order.Id,
				items[i].ProductId,
				items[i].Qty,
				items[i].Price,
			}, nil
		}),
	)
	if err != nil {
		return
	}

	// delete user cart items
	if _, err = tx.Exec(ctx, "delete from cart_items where user_id=$1", userId); err != nil {
		return
	}

	// update products stock
	b := &pgx.Batch{}
	for _, item := range items {
		q := `update products set stock=stock-$1 where id=$2`
		b.Queue(q, item.Qty, item.ProductId)
	}
	if err = tx.SendBatch(ctx, b).Close(); err != nil {
		return
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return
}
