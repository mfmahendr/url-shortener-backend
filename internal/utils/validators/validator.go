package validators

import (
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func Init() {
	Validate = validator.New()
	
	registerCustomValidations()
}

func registerCustomValidations() {
	Validate.RegisterValidation("short_id", CustomIDFormat)
}