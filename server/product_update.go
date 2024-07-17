package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	prd "sypchal/product"
	"sypchal/validation"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type ProductUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ImageUrl    string `json:"image_url,omitempty"`
	Category    string `json:"category,omitempty"`
	Stock       int    `json:"stock,omitempty"`
	Price       int    `json:"price,omitempty"`
}

func (s *ServerDependency) ProductUpdate(w http.ResponseWriter, r *http.Request) {
	productId, _ := strconv.Atoi(chi.URLParam(r, "id"))
	requestBody := ProductUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.ErrorResponse(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	product, err := s.productDomain.UpdateProductById(
		r.Context(),
		productId,
		prd.UpdateProductRequest(requestBody),
	)
	if err != nil {
		log.Error().Err(err).Msg("update product")

		var ve *validation.ValidationErrors
		if errors.As(err, &ve) {
			w.WriteHeader(http.StatusBadRequest)
			s.ErrorResponse(w, http.StatusBadRequest, "validation error", ve.Transform())
			return
		}

		if errors.Is(err, prd.ErrProductNotFound) {
			w.WriteHeader(http.StatusBadRequest)
			s.ErrorResponse(w, http.StatusBadRequest, "product not found", nil)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		s.ErrorResponse(w, http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.DataResponse(w, product)
}
