package api

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func JSONError(w http.ResponseWriter, status int, message string, details string) {
	JSON(w, status, ErrorResponse{
		Error:   message,
		Code:    status,
		Details: details,
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusBadRequest, message, "")
}

func NotFound(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusNotFound, message, "")
}

func InternalError(w http.ResponseWriter, message string, details string) {
	JSONError(w, http.StatusInternalServerError, message, details)
}

func Unauthorized(w http.ResponseWriter, message string) {
	JSONError(w, http.StatusUnauthorized, message, "")
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
