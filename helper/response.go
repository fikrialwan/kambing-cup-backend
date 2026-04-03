package helper

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitzero"`
	ErrorCode string      `json:"error_code,omitzero"`
	Message   string      `json:"message"`
}

func WriteResponse(w http.ResponseWriter, statusCode int, success bool, data interface{}, errorCode string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success:   success,
		Data:      data,
		ErrorCode: errorCode,
		Message:   message,
	})
}
