package dto

type ShortenerRequest struct {
	URL      string `json:"url"`
	CustomID string `json:"custom_id"`
}

type ShortenerResponse struct {
	ShortID string `json:"short_id"`
}
