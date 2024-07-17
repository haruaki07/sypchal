package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

type ValidationErrors struct {
	*validator.ValidationErrors
}

func (e ValidationErrors) Transform() map[string]string {
	errors := map[string]string{}

	for _, err := range *e.ValidationErrors {
		if err.Tag() == "required" {
			errors[err.Field()] = fmt.Sprintf("%s is required", err.Field())
		} else if err.Tag() == "email" {
			errors[err.Field()] = fmt.Sprintf("%s must be a valid email", err.Field())
		} else {
			errors[err.Field()] = fmt.Sprintf(
				"%s %s %s",
				err.Field(),
				translation[err.Tag()],
				err.Param(),
			)
		}
	}

	return errors
}

var translation = map[string]interface{}{
	"gte": "must be greater than or equal to",
	"lte": "must be less than or equal to",
	"gt":  "must be greater than",
	"lt":  "must be less than",
	"url": "must be a valid url",
	"min": "must be at least",
	"max": "can not be more than",
}

func NewValidator() *Validator {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &Validator{validate}
}

func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		if errs, ok := err.(validator.ValidationErrors); ok {
			return &ValidationErrors{ValidationErrors: &errs}
		}
	}

	return nil
}
