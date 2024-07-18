package cart

import "errors"

var ErrProductNotFound = errors.New("product not found")
var ErrProductOutOfStock = errors.New("product out of stock")
