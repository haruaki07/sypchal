package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sypchal/cart"
	"sypchal/validation"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

type CartUpdateItemRequest struct {
	Qty int `json:"qty"`
}

func (s *ServerDependency) CartUpdateItem(w http.ResponseWriter, r *http.Request) {
	itemId, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var requestBody CartUpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		s.Response(w, r).Status(http.StatusBadRequest).
			Error(http.StatusBadRequest, "invalid request body", nil)
		return
	}

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

	count, err := s.cartDomain.UpdateCartItem(r.Context(), cart.UpdateCartItemRequest{
		UserId: userId,
		ItemId: itemId,
		Qty:    requestBody.Qty,
	})
	if err != nil {
		log.Error().Err(err).Msg("update cart item")

		var ve *validation.ValidationErrors
		if errors.As(err, &ve) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "validation error", ve.Transform())
			return
		}

		if errors.Is(err, cart.ErrProductNotFound) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "product not found", nil)
			return
		}

		if errors.Is(err, cart.ErrProductOutOfStock) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "product out of stock", nil)
			return
		}

		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Data(map[string]interface{}{"count": count})
}
