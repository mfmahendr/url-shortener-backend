package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
)

func (c *URLController) FetchBlacklistedDomains(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	domains, err := c.blacklistManager.ListBlacklistedDomains(r.Context())
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to retrieve blacklist: "+err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(domains); err != nil {
		http.Error(w, "Failed to encode response", 500)
		return
	}
}

func (c *URLController) AddToBlacklist(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req dto.BlacklistDomain
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Domain == "" {
		http.Error(w, "Invalid request", 400)
		return
	}
	err := c.blacklistManager.BlacklistDomain(r.Context(), req.Domain)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to add domain to blacklist: "+err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "added", "domain": req.Domain}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", 500)
		return
	}
}

func (c *URLController) RemoveFromBlacklist(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	domain := ps.ByName("domain")
	if domain == "" {
		http.Error(w, "Missing domain", 400)
		return
	}
	err := c.blacklistManager.UnblacklistDomain(r.Context(), domain)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		http.Error(w, "Failed to remove domain to blacklist: "+err.Error(), statusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "removed", "domain": domain}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", 500)
		return
	}
}
