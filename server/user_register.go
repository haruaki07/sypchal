package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"sypchal/user"
	"sypchal/validation"

	"github.com/rs/zerolog/log"
)

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

func (s *ServerDependency) UserRegister(w http.ResponseWriter, r *http.Request) {
	requestBody := UserRegisterRequest{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		s.Response(w, r).Status(http.StatusBadRequest).
			Error(http.StatusBadRequest, "invalid request body", nil)

		return
	}

	err := s.userDomain.CreateUser(r.Context(), user.CreateUserRequest{
		Email:    requestBody.Email,
		Password: requestBody.Password,
		FullName: requestBody.FullName,
	})
	if err != nil {
		log.Error().Err(err).Msg("create user")

		var ve *validation.ValidationErrors
		if errors.As(err, &ve) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "validation error", ve.Transform())
			return
		}

		if errors.Is(err, user.ErrEmailAlreadyExists) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "email is already registered", nil)
			return
		}

		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Status(http.StatusCreated).End()
}
