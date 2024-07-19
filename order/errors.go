package order

import "errors"

var ErrItemOutOfStock = errors.New("item out of stock")
var ErrOrderNotFound = errors.New("order not found")
var ErrPayAmountNotMatch = errors.New("pay amount not match")
var ErrPaymentIdMismatch = errors.New("pay_id mismatch")
var ErrOrderIsPaid = errors.New("order is paid")
