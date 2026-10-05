package rest

import "github.com/go-playground/validator/v10"

var validators = validator.New(
	validator.WithRequiredStructEnabled(),
)
