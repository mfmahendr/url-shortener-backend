package validators

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	Validate *validator.Validate
	once     sync.Once
)

func Init() {
	once.Do(func() {
		Validate = validator.New()

		registerCustomValidations()
	})
}

func registerCustomValidations() {
	Validate.RegisterValidation("short_id", CustomIDFormat)
}