package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"sypchal/user"
	"sypchal/validation"

	"github.com/rs/zerolog/log"
)

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *ServerDependency) UserLogin(w http.ResponseWriter, r *http.Request) {
	requestBody := UserLoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.ErrorResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	accessToken, err := s.userDomain.Authenticate(r.Context(), user.AuthenticateRequest{
		Email:    requestBody.Email,
		Password: requestBody.Password,
	})
	if err != nil {
		log.Error().Err(err).Msg("authenticate")

		var ve *validation.ValidationErrors
		if errors.As(err, &ve) {
			w.WriteHeader(http.StatusBadRequest)
			s.ErrorResponse(w, http.StatusBadRequest, "validation error", ve.Transform())
			return
		}

		if errors.Is(err, user.ErrWrongEmailOrPassword) {
			w.WriteHeader(http.StatusBadRequest)
			s.ErrorResponse(w, http.StatusBadRequest, "wrong email or password", nil)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		s.ErrorResponse(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.DataResponse(w, map[string]string{"access_token": accessToken})
}
