package order

import (
	"context"
	"errors"
	"sypchal/validation"
	"time"

	"github.com/jackc/pgx/v5"
)

type OrderDomain struct {
	db        *pgx.Conn
	validator *validation.Validator
}

func NewOrderDomain(db *pgx.Conn, validator *validation.Validator) (*OrderDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if validator == nil {
		return nil, errors.New("validator is nil")
	}

	return &OrderDomain{db, validator}, nil
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
	PayId      string     `json:"pay_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

type Payment struct {
	Id        int        `json:"id"`
	OrderId   int        `json:"order_id"`
	UserId    int        `json:"user_id"`
	ProofUrl  string     `json:"proof_url"`
	Amount    int        `json:"amount"`
	Method    string     `json:"method"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type CartItem struct {
	TotalPrice   int
	ProductId    int
	ProductStock int
	Qty          int
	Price        int
}

func (o *OrderDomain) PlaceOrder(ctx context.Context, userId int) (order *Order, err error) {
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
	order = &Order{}
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
		q := `update products set stock=stock-$1,updated_at=now() where id=$2`
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

type PayOrderRequest struct {
	PayId    string
	OrderId  int    `json:"order_id" validate:"required"`
	ProofUrl string `json:"proof_url" validate:"required,http_url"`
	Amount   int    `json:"amount" validate:"required"`
	Method   string `json:"method" validate:"required"`
}

func (o *OrderDomain) PayOrder(ctx context.Context, userId int, req PayOrderRequest) (payment *Payment, err error) {
	if err = o.validator.ValidateStruct(req); err != nil {
		return
	}

	tx, err := o.db.Begin(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	var totalPrice int
	var payIdDb string
	var orderStatus string
	err = tx.QueryRow(ctx, "select total_price,pay_id,status from orders where id=$1", req.OrderId).
		Scan(&totalPrice, &payIdDb, &orderStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrOrderNotFound
		}

		return
	}

	if orderStatus != OrderStatusUnpaid {
		err = ErrOrderIsPaid
		return
	}

	if payIdDb != req.PayId {
		err = ErrPaymentIdMismatch
		return
	}

	if totalPrice > req.Amount {
		err = ErrPayAmountNotMatch
		return
	}

	payment = &Payment{}
	err = tx.QueryRow(
		ctx,
		`insert into payments (order_id,user_id,proof_url,amount,method)
		values ($1,$2,$3,$4,$5) returning *`,
		req.OrderId,
		userId,
		req.ProofUrl,
		req.Amount,
		req.Method,
	).Scan(
		&payment.Id,
		&payment.OrderId,
		&payment.UserId,
		&payment.ProofUrl,
		&payment.Amount,
		&payment.Method,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return
	}

	_, err = tx.Exec(ctx, "update orders set status=$1,updated_at=now() where id=$2", OrderStatusPaid, req.OrderId)
	if err != nil {
		return
	}

	if err = tx.Commit(ctx); err != nil {
		return
	}

	return
}
