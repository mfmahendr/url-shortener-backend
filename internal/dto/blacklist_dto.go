package dto

type BlacklistDomain struct {
	Domain    string    `json:"domain,omitempty" validate:"required,hostname"`
}