package user

import (
	"context"
	"errors"
	"fmt"
	"sypchal/validation"
	"time"

	"github.com/go-chi/jwtauth/v5"
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
	Jwt       *jwtauth.JWTAuth
}

func NewUserDomain(db *pgx.Conn, validator *validation.Validator, jwtSecret string) (*UserDomain, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	jwt := jwtauth.New("HS256", []byte(jwtSecret), nil)

	return &UserDomain{db, validator, jwt}, nil
}

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
}

func (r CreateUserRequest) validate(validator *validation.Validator) error {
	return validator.ValidateStruct(r)
}

func (u *UserDomain) CreateUser(ctx context.Context, req CreateUserRequest) error {
	if err := req.validate(u.validator); err != nil {
		return err
	}

	var exists int
	err := u.db.QueryRow(ctx, "select count(*) from users where email = $1", req.Email).Scan(&exists)
	if err != nil {
		return err
	}

	if exists > 0 {
		return ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = u.db.Exec(
		ctx,
		"insert into users(email, password, full_name) values($1, $2, $3)",
		req.Email,
		string(hash),
		req.FullName,
	)
	if err != nil {
		return err
	}

	return nil
}

type AuthenticateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (u *UserDomain) Authenticate(ctx context.Context, req AuthenticateRequest) (accessToken string, err error) {
	user := User{}
	err = u.db.QueryRow(
		ctx,
		"select id, email, password from users where email = $1",
		req.Email,
	).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrWrongEmailOrPassword
			return
		}

		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		err = ErrWrongEmailOrPassword
		return
	}

	_, accessToken, err = u.Jwt.Encode(map[string]interface{}{
		"uid": user.Id,
		"exp": time.Now().Add(time.Hour * 1).Unix(), // an hour
	})
	if err != nil {
		err = fmt.Errorf("signing token: %w", err)
		return
	}

	return
}
