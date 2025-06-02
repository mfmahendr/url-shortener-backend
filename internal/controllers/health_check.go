package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (c *URLController) HealthCheck(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"status":  "ok",
		"message": "Service is running",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}