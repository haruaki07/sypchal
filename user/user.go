package user

import (
	"context"
	"fmt"
	"sypchal/validation"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id        string     `json:"id"`
	Email     string     `json:"email"`
	Password  string     `json:"password"`
	FullName  string     `json:"full_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type UserDomain struct {
	db        *pgx.Conn
	validator *validation.Validator
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
}

func (r CreateUserRequest) validate(validator *validation.Validator) error {
	return validator.ValidateStruct(r)
}

func NewUserDomain(db *pgx.Conn, validator *validation.Validator) (*UserDomain, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	return &UserDomain{db, validator}, nil
}

func (u *UserDomain) CreateUser(ctx context.Context, req CreateUserRequest) (User, error) {
	if err := req.validate(u.validator); err != nil {
		return User{}, err
	}

	var exists int
	err := u.db.QueryRow(ctx, "select count(*) from users where email = $1", req.Email).Scan(&exists)
	if err != nil {
		return User{}, err
	}

	if exists > 0 {
		return User{}, ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	user := User{}
	err = u.db.QueryRow(
		ctx,
		"insert into users(email, password, full_name) values($1, $2, $3) returning *",
		req.Email,
		string(hash),
		req.FullName,
	).Scan(&user.Id, &user.Email, &user.Password, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, nil
	}

	return user, nil
}
