package server

import (
	"errors"
	"net/http"
	"strconv"
	"sypchal/cart"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

func (s *ServerDependency) CartDeleteItem(w http.ResponseWriter, r *http.Request) {
	itemId, _ := strconv.Atoi(chi.URLParam(r, "id"))

	_, payload, err := jwtauth.FromContext(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("get jwt payload")
		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	userId, err := strconv.Atoi(payload["uid"].(string))
	if err != nil {
		log.Error().Err(err).Msg("atoi")
		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	count, err := s.cartDomain.DeleteCartItem(r.Context(), cart.DeleteCartItemRequest{
		UserId: userId,
		ItemId: itemId,
	})
	if err != nil {
		log.Error().Err(err).Msg("delete cart item")

		if errors.Is(err, cart.ErrCartItemNotFound) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "cart item not found", nil)
			return
		}

		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Status(http.StatusCreated).
		Data(map[string]interface{}{"count": count})
}
