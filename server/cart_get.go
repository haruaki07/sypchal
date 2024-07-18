package server

import (
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

func (s *ServerDependency) CartGet(w http.ResponseWriter, r *http.Request) {
	_, payload, err := jwtauth.FromContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.ErrorResponse(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	userId, err := strconv.Atoi(payload["uid"].(string))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.ErrorResponse(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	cart, err := s.cartDomain.GetUserCart(r.Context(), userId)
	if err != nil {
		log.Error().Err(err).Msg("get user cart")

		w.WriteHeader(http.StatusInternalServerError)
		s.ErrorResponse(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.DataResponse(w, cart)
}
