package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sypchal/order"
	"sypchal/validation"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog/log"
)

type OrderPayRequest struct {
	OrderId  int    `json:"order_id"`
	ProofUrl string `json:"proof_url"`
	Amount   int    `json:"amount"`
	Method   string `json:"method"`
}

func (s *ServerDependency) OrderPay(w http.ResponseWriter, r *http.Request) {
	requestBody := OrderPayRequest{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		s.Response(w, r).Status(http.StatusBadRequest).
			Error(http.StatusBadRequest, "invalid request body", nil)
		return
	}

	payId := chi.URLParam(r, "pay_id")

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

	payment, err := s.orderDomain.PayOrder(r.Context(), userId, order.PayOrderRequest{
		PayId:    payId,
		OrderId:  requestBody.OrderId,
		ProofUrl: requestBody.ProofUrl,
		Amount:   requestBody.Amount,
		Method:   requestBody.Method,
	})
	if err != nil {
		log.Error().Err(err).Msg("pay order")

		var ve *validation.ValidationErrors
		if errors.As(err, &ve) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "validation error", ve.Transform())
			return
		}

		if errors.Is(err, order.ErrOrderNotFound) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "order not found", nil)
			return
		}

		if errors.Is(err, order.ErrOrderIsPaid) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "order has been paid", nil)
			return
		}

		if errors.Is(err, order.ErrPaymentIdMismatch) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "pay id mismatch", nil)
			return
		}

		if errors.Is(err, order.ErrPayAmountNotMatch) {
			s.Response(w, r).Status(http.StatusBadRequest).
				Error(http.StatusBadRequest, "pay amount not match", nil)
			return
		}

		s.Response(w, r).Status(http.StatusInternalServerError).
			Error(http.StatusInternalServerError, "internal server error", nil)
		return
	}

	s.Response(w, r).Data(payment)
}
