package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

func CustomIDFormat(fl validator.FieldLevel) bool {
		return regexp.MustCompile(`^[a-zA-Z0-9]{3,30}$`).MatchString(fl.Field().String())
}