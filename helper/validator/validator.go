package validator

import "github.com/go-playground/validator/v10"

func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

type Validator struct {
	validate *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

func (t *Validator) Validator() *Validator {
	return t
}
