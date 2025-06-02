package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (c *URLController) Home(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the URL Shortener API",
	})
}