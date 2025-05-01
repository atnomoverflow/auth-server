package validation

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// register new validation
	// validate.RegisterValidation("passwd", validatePassword)
}

func Validate(body io.ReadCloser, dto any) error {
	err := json.NewDecoder(body).Decode(dto)
	if err != nil {
		return err
	}
	err = validate.Struct(dto)
	return err
}
