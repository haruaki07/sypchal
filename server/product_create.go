package server

import (
	"encoding/json"
	"errors"
	"net/http"
	prd "sypchal/product"
	"sypchal/validation"

	"github.com/rs/zerolog/log"
)

type ProductCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageUrl    string `json:"image_url"`
	Category    string `json:"category"`
	Stock       int    `json:"stock"`
	Price       int    `json:"price"`
}

func (s *ServerDependency) ProductCreate(w http.ResponseWriter, r *http.Request) {
	requestBody := ProductCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		s.Response(w, r).Status(http.StatusBadRequest).
			Error(http.StatusBadRequest, "invalid request body", nil)

		return
	}

	product, err := s.productDomain.CreateProduct(r.Context(), prd.CreateProductRequest(requestBody))
	if err != nil {
		log.Error().Err(err).Msg("create product")

		var ve *validation.ValidationErrors
		if errors.As(err, &ve) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "validation error", ve.Transform())
			return
		}

		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Status(http.StatusCreated).
		Data(product)
}
