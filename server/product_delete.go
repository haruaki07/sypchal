package server

import (
	"errors"
	"net/http"
	"strconv"
	prd "sypchal/product"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func (s *ServerDependency) ProductDelete(w http.ResponseWriter, r *http.Request) {
	productId, _ := strconv.Atoi(chi.URLParam(r, "id"))

	err := s.productDomain.DeleteProductById(
		r.Context(),
		productId,
	)
	if err != nil {
		log.Error().Err(err).Msg("delete product by id")

		if errors.Is(err, prd.ErrProductNotFound) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "product not found", nil)
			return
		}

		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Status(http.StatusNoContent).End()
}
