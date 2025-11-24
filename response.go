package fluxo

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, data interface{}) error {
	return JSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data interface{}) error {
	return JSON(w, http.StatusCreated, data)
}

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}