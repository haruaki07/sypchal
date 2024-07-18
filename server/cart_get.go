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
		log.Error().Err(err).Msg("get jwt payload")

		s.Response(w, r).
			Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	userId, err := strconv.Atoi(payload["uid"].(string))
	if err != nil {
		log.Error().Err(err).Msg("atoi")

		s.Response(w, r).
			Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	cart, err := s.cartDomain.GetUserCart(r.Context(), userId)
	if err != nil {
		log.Error().Err(err).Msg("get user cart")

		s.Response(w, r).
			Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Data(cart)
}
