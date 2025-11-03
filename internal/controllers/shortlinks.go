package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mfmahendr/url-shortener-backend/internal/dto"
	"github.com/mfmahendr/url-shortener-backend/internal/utils"
)

func (c *URLController) GetShortlinks(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()
	createdBy, ok := ctx.Value(utils.UserKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// queries
	var isPrivateQ string
	paginationQ := new(dto.PaginationQuery)
	if isPrivateQ = r.URL.Query().Get("is_private"); isPrivateQ != "" {
		isPrivateQ = "all"
	}
	parsePaginationQuery(r, paginationQ)
	shortlinksQ := &dto.UserLinksQuery{
		IsPrivate:       isPrivateQ,
		PaginationQuery: *paginationQ,
	}

	linksReq := &dto.UserLinksRequest{
		CreatedBy:      createdBy,
		UserLinksQuery: *shortlinksQ,
	}

	resp, err := c.shortenService.GetUserLinks(ctx, *linksReq)
	if err != nil {
		statusCode := mapErrorToStatusCode(err)
		if statusCode != http.StatusNotFound {
			http.Error(w, "Failed to get users' shortlinks: "+err.Error(), statusCode)
		}
	}
	resp.CreatedBy = linksReq.CreatedBy

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
