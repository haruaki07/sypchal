package cart

import "errors"

var ErrProductNotFound = errors.New("product not found")
var ErrProductOutOfStock = errors.New("product out of stock")
var ErrCartItemNotFound = errors.New("cart item not found")
