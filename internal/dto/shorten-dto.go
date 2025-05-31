package dto

type ShortenerRequest struct {
	URL      string `json:"url" validate:"required,url"`
	CustomID string `json:"custom_id" validate:"omitempty,short_id"`
}

type ShortenerResponse struct {
	ShortID string `json:"short_id"`
}
