package dto

type ShortenRequest struct {
	URL       string `json:"url" validate:"required,url"`
	CustomID  string `json:"custom_id" validate:"omitempty,short_id"`
	IsPrivate bool   `json:"is_private"`
}

type ShortenResponse struct {
	ShortID string `json:"short_id"`
}
