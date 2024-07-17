package user

import "errors"

var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrWrongEmailOrPassword = errors.New("wrong email or password")
