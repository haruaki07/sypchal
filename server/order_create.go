package server

import (
	"errors"
	"net/http"
	"strconv"
	"sypchal/order"

	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

func (s *ServerDependency) OrderCreate(w http.ResponseWriter, r *http.Request) {
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

	userOrder, err := s.orderDomain.PlaceOrder(r.Context(), userId)
	if err != nil {
		log.Error().Err(err).Msg("place order")

		if errors.Is(err, order.ErrItemOutOfStock) {
			s.Response(w, r).
				Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "item out of stock", nil)
			return
		}

		s.Response(w, r).
			Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Status(http.StatusCreated).Data(userOrder)
}
