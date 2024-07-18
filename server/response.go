package server

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"
)

type ErrorField struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

type CommonResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *ErrorField `json:"error,omitempty"`
}

type response struct {
	w http.ResponseWriter
	r *http.Request
}

func (s *ServerDependency) Response(w http.ResponseWriter, r *http.Request) *response {
	return &response{w, r}
}

type statusCtxKey string

var StatusCtxKey = statusCtxKey("http.status")

func (r *response) Status(status int) *response {
	*r.r = *(r.r).WithContext(context.WithValue(r.r.Context(), StatusCtxKey, status))
	return r
}

func (r *response) End() {
	if status, ok := r.r.Context().Value(StatusCtxKey).(int); ok {
		r.w.WriteHeader(status)
	}
}

func (r *response) Json(body CommonResponse) {
	r.w.Header().Set("Content-Type", "application/json")

	if status, ok := r.r.Context().Value(StatusCtxKey).(int); ok {
		r.w.WriteHeader(status)
	}

	_ = json.NewEncoder(r.w).Encode(body)
}

func (r *response) Data(data interface{}) {
	r.Json(CommonResponse{
		Data:  data,
		Error: nil,
	})
}

func (r *response) Error(code int, message string, errors interface{}) {
	r.Json(CommonResponse{
		Error: &ErrorField{
			Code:    code,
			Message: message,
			Errors:  errors,
		},
	})
}
