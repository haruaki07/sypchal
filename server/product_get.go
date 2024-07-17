package server

import (
	"errors"
	"net/http"
	"strconv"
	prd "sypchal/product"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (s *ServerDependency) ProductGet(w http.ResponseWriter, r *http.Request) {
	productId, _ := strconv.Atoi(chi.URLParam(r, "id"))

	product, err := s.productDomain.GetProductById(
		r.Context(),
		productId,
	)
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

	s.DataResponse(w, product)
}
