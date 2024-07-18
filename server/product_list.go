package server

import (
	"net/http"
	"strconv"
	"sypchal/product"

	"github.com/rs/zerolog/log"
)

func (s *ServerDependency) ProductList(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	offset := limit * (page - 1)

	res, err := s.productDomain.GetProducts(r.Context(), product.GetProductsRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Error().Err(err).Msg("get products")

		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Data(res)
}
