package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
)

func (c *URLController) FetchBlacklistItems(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	items, err := c.blacklistManager.ListBlacklisted(r.Context())
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to retrieve blacklist: "+err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (c *URLController) AddToBlacklist(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req dto.BlacklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Type == "" || req.Value == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var err error
	switch req.Type {
	case "domain":
		err = c.blacklistManager.BlacklistDomain(r.Context(), req.Value)
	case "url":
		err = c.blacklistManager.BlacklistURL(r.Context(), req.Value)
	default:
		http.Error(w, "Invalid type: must be 'domain' or 'url'", http.StatusBadRequest)
		return
	}

	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to add to blacklist: "+err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"status": "added", "type": req.Type, "value": req.Value}
	_ = json.NewEncoder(w).Encode(resp)
}


func (c *URLController) RemoveFromBlacklist(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	blacklistValue := r.URL.Query().Get("value")
	blacklistType := r.URL.Query().Get("type")

	if blacklistValue == "" || blacklistType == "" {
		http.Error(w, "Missing value or type", http.StatusBadRequest)
		return
	}

	var err error
	switch blacklistType {
	case "domain":
		err = c.blacklistManager.UnblacklistDomain(r.Context(), blacklistValue)
	case "url":
		err = c.blacklistManager.UnblacklistURL(r.Context(), blacklistValue)
	default:
		http.Error(w, "Invalid type: must be 'domain' or 'url'", http.StatusBadRequest)
		return
	}

	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to remove from blacklist: "+err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "removed", "type": blacklistType, "value": blacklistValue}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", 500)
		return
	}
}