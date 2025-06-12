package utils

type contextKey string

const (
	UserKey contextKey = "user"
	ExportFormatKey contextKey = "export_format"
)