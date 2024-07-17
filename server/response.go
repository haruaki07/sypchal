package server

import (
	"encoding/json"
	"net/http"
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

func (s *ServerDependency) ResponseJson(w http.ResponseWriter, data CommonResponse) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *ServerDependency) DataResponse(w http.ResponseWriter, data interface{}) {
	s.ResponseJson(w, CommonResponse{
		Data:  data,
		Error: nil,
	})
}

func (s *ServerDependency) ErrorResponse(w http.ResponseWriter, code int, message string, errors interface{}) {
	s.ResponseJson(w, CommonResponse{
		Error: &ErrorField{
			Code:    code,
			Message: message,
			Errors:  errors,
		},
	})
}
