package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sypchal/validation"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var testDb *pgx.Conn

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %v", err)
	}

	resource, err := pool.Run("postgres", "alpine", []string{"POSTGRES_PASSWORD=postgres"})
	if err != nil {
		log.Fatalf("could not start resource: %v", err)
	}

	ctx := context.Background()
	var connStr string
	trial := 0
	if err := pool.Retry(func() error {
		trial++
		log.Printf("trying to connect to postgres.. %d\n", trial)
		connStr = fmt.Sprintf(
			"postgres://postgres:postgres@localhost:%s/postgres",
			resource.GetPort("5432/tcp"),
		)
		testDb, err = pgx.Connect(ctx, connStr)
		if err != nil {
			return fmt.Errorf("failed to connect test database, retrying: %v", err)
		}
		return nil
	}); err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}

	// migrate db
	log.Println("migrating db..")
	if err := dbMigrate(ctx, connStr); err != nil {
		log.Fatalf("could not migrate db: %s", err)
	}

	exitCode := m.Run()

	log.Println("purging postgres..")
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %v", err)
	}

	os.Exit(exitCode)
}

func TestNewUserDomain(t *testing.T) {
	validator := validation.NewValidator()
	jwtSecret := "testing"

	t.Run("all dependencies set", func(t *testing.T) {
		userDomain, err := NewUserDomain(testDb, validator, jwtSecret)
		if err != nil {
			t.Errorf("NewUserDomain failed: %v", err)
		}
		if userDomain == nil {
			t.Error("NewUserDomain returned nil userDomain")
		}
	})

	t.Run("nil db", func(t *testing.T) {
		userDomain, err := NewUserDomain(nil, validator, jwtSecret)
		if err == nil {
			t.Error("NewUserDomain returned nil error with nil db")
		}
		if userDomain != nil {
			t.Error("NewUserDomain returned nil userDomain with nil db")
		}
	})

	t.Run("nil validator", func(t *testing.T) {
		userDomain, err := NewUserDomain(testDb, nil, jwtSecret)
		if err == nil {
			t.Error("NewUserDomain returned nil error with nil validator")
		}
		if userDomain != nil {
			t.Error("NewUserDomain returned nil userDomain with nil validator")
		}
	})

	t.Run("empty jwtSecret", func(t *testing.T) {
		userDomain, err := NewUserDomain(testDb, validator, "")
		if err == nil {
			t.Error("NewUserDomain returned nil error with empty jwtSecret")
		}
		if userDomain != nil {
			t.Error("NewUserDomain returned nil userDomain with empty jwtSecret")
		}
	})
}

func TestCreateUser(t *testing.T) {
	validator := validation.NewValidator()
	jwtSecret := "testing"
	userDomain, err := NewUserDomain(testDb, validator, jwtSecret)
	if err != nil {
		t.Fatalf("creating user domain instance: %v", err)
	}

	t.Run("empty request fields", func(t *testing.T) {
		req := CreateUserRequest{
			Email:    "",
			Password: "",
			FullName: "",
		}

		err := userDomain.CreateUser(context.Background(), req)
		if err == nil {
			t.Errorf("expecting an error, got nil instead: %v", err)
		}

		var ve *validation.ValidationErrors
		if !errors.As(err, &ve) {
			t.Errorf("expecting validation error, got %v", err)
		}

		if len(*ve.ValidationErrors) != 3 {
			t.Errorf("expecting 3 errors, got %d", len(*ve.ValidationErrors))
		}
	})

	t.Run("already exists email", func(t *testing.T) {
		defer (func() {
			_, err := testDb.Exec(context.Background(), "delete from users")
			if err != nil {
				t.Fatalf("failed to clean up users table: %v", err)
			}
		})()

		_, err := testDb.Exec(
			context.Background(),
			"insert into users(email,password,full_name) values('john@mail.com', '123', 'John Doe')",
		)
		if err != nil {
			t.Fatalf("failed to seed users table: %v", err)
		}

		req := CreateUserRequest{
			Email:    "john@mail.com",
			Password: "123456",
			FullName: "John Doe",
		}

		err = userDomain.CreateUser(context.Background(), req)
		if err == nil {
			t.Errorf("expecting an error, got nil instead: %v", err)
		}

		if !errors.Is(err, ErrEmailAlreadyExists) {
			t.Errorf("expecting email exists error, got %v", err)
		}
	})

	t.Run("valid user", func(t *testing.T) {
		defer (func() {
			_, err := testDb.Exec(context.Background(), "delete from users")
			if err != nil {
				t.Fatalf("failed to clean up users table: %v", err)
			}
		})()

		req := CreateUserRequest{
			Email:    "john.doe@mail.com",
			Password: "123456",
			FullName: "John Doe",
		}

		err := userDomain.CreateUser(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		var id int
		err = testDb.
			QueryRow(context.Background(), "select id from users where email=$1", req.Email).
			Scan(&id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				t.Error("expecting new user id, got no rows instead")
			}

			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestAuthenticate(t *testing.T) {
	validator := validation.NewValidator()
	jwtSecret := "testing"
	userDomain, err := NewUserDomain(testDb, validator, jwtSecret)
	if err != nil {
		t.Fatalf("creating user domain instance: %v", err)
	}

	defer (func() {
		_, err := testDb.Exec(context.Background(), "delete from users")
		if err != nil {
			t.Fatalf("failed to clean up users table: %v", err)
		}
	})()

	_, err = testDb.Exec(
		context.Background(),
		"insert into users(email,password,full_name) values('john.doe@mail.com', $1, 'John Doe')",
		"$2a$10$QPe3VMaeWrIGwvGZHdU2BeExPEv9O86PUvpD0/0vbpT2FwgA4Yqpm", // 123456
	)
	if err != nil {
		t.Fatalf("failed to seed users table: %v", err)
	}

	t.Run("invalid request", func(t *testing.T) {
		req := AuthenticateRequest{
			Email:    "",
			Password: "",
		}

		accessToken, err := userDomain.Authenticate(context.Background(), req)
		if err == nil {
			t.Errorf("expecting an error, got nil instead: %v", err)
		}

		var ve *validation.ValidationErrors
		if !errors.As(err, &ve) {
			t.Errorf("expecting validation error, got %v", err)
		}

		if len(*ve.ValidationErrors) != 2 {
			t.Errorf("expecting 2 errors, got %d", len(*ve.ValidationErrors))
		}

		if accessToken != "" {
			t.Errorf("expecting empty accessToken, got non empty instead: %s", accessToken)
		}
	})

	t.Run("wrong email", func(t *testing.T) {
		req := AuthenticateRequest{
			Email:    "john.d@mail.com",
			Password: "123456",
		}

		accessToken, err := userDomain.Authenticate(context.Background(), req)
		if err == nil {
			t.Errorf("expecting an error, got nil instead: %v", err)
		}

		if !errors.Is(err, ErrWrongEmailOrPassword) {
			t.Errorf("expecting wrong email or password error, got %v", err)
		}

		if accessToken != "" {
			t.Errorf("expecting empty accessToken, got non empty instead: %s", accessToken)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		req := AuthenticateRequest{
			Email:    "john.doe@mail.com",
			Password: "12345",
		}

		accessToken, err := userDomain.Authenticate(context.Background(), req)
		if err == nil {
			t.Errorf("expecting an error, got nil instead: %v", err)
		}

		if !errors.Is(err, ErrWrongEmailOrPassword) {
			t.Errorf("expecting wrong email or password error, got %v", err)
		}

		if accessToken != "" {
			t.Errorf("expecting empty accessToken, got non empty instead: %s", accessToken)
		}
	})

	t.Run("valid user", func(t *testing.T) {
		req := AuthenticateRequest{
			Email:    "john.doe@mail.com",
			Password: "123456",
		}

		accessToken, err := userDomain.Authenticate(context.Background(), req)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if accessToken == "" {
			t.Error("expecting jwt string accessToken, got empty string instead")
		}
	})
}

func dbMigrate(ctx context.Context, connStr string) error {
	db, err := goose.OpenDBWithDriver("postgres", connStr)
	if err != nil {
		return fmt.Errorf("goose: failed to open DB: %v", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get cwd: %v", err)
	}

	goose.SetLogger(goose.NopLogger()) // surpress log
	return goose.RunContext(ctx, "up", db, filepath.Join(filepath.Dir(dir), "migrations"))
}
