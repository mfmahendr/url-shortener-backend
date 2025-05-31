package utils

var BlacklistedCustomIDs  = map[string]bool{
	"admin":    true,
	"api":      true,
	"shorten":  true,
	"login":    true,
	"logout":   true,
	"dashboard": true,
}