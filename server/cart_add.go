package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sypchal/cart"
	prd "sypchal/product"
	"sypchal/validation"

	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

type CartAddRequest struct {
	ProductId int `json:"product_id"`
	Qty       int `json:"qty"`
}

func (s *ServerDependency) CartAdd(w http.ResponseWriter, r *http.Request) {
	var requestBody CartAddRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.ErrorResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

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

	if requestBody.ProductId == 0 {
		w.WriteHeader(http.StatusBadRequest)
		s.ErrorResponse(w, http.StatusBadRequest, "product_id is required", nil)
		return
	}

	product, err := s.productDomain.GetProductById(r.Context(), requestBody.ProductId)
	if err != nil {
		log.Error().Err(err).Msg("get product by id")

		if errors.Is(err, prd.ErrProductNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			s.ErrorResponse(w, http.StatusBadRequest, "product not found", nil)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		s.ErrorResponse(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	count, err := s.cartDomain.AddCartItem(r.Context(), cart.AddCartItemRequest{
		UserId:    userId,
		ProductId: product.Id,
		Qty:       requestBody.Qty,
		Price:     product.Price,
	})
	if err != nil {
		log.Error().Err(err).Msg("add cart item")

		var ve *validation.ValidationErrors
		if errors.As(err, &ve) {
			w.WriteHeader(http.StatusBadRequest)
			s.ErrorResponse(w, http.StatusBadRequest, "validation error", ve.Transform())
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		s.ErrorResponse(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	w.WriteHeader(http.StatusCreated)
	s.DataResponse(w, map[string]interface{}{
		"count": count,
	})
}
