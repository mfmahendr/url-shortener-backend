package dto

type BlacklistItemRequest struct {
	Type  string `json:"type"`  // "domain" or "url"
	Value string `json:"value"`
}