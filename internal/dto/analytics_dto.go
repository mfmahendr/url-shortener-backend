package dto

import "github.com/mfmahendr/url-shortener-backend/internal/models"

type AnalyticsDTO struct {
	ShortID     string            `json:"short_id"`
	TotalClicks int               `json:"total_clicks"`
	Clicks      []models.ClickLog `json:"clicks"`
}
