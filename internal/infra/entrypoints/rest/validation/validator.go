package validation

import (
	"context"
	"errors"
	"fmt"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	english "github.com/go-playground/locales/en"
	en_translation "github.com/go-playground/validator/v10/translations/en"
)

type ValidationErrors struct {
	Message string            `json:"message"`
	Errors  []ValidationError `json:"errors,omitempty"`
}
type ValidationError struct {
	Message       string `json:"message"`
	FieldName     string `json:"field_name"`
	RejectedValue any    `json:"rejected_value"`
}

func (e ValidationErrors) Error() string {
	return fmt.Sprintf("%d error(s): %s", len(e.Errors), e.Message)
}
func (e ValidationError) Error() string {
	return fmt.Sprintf("Error on field %s: %s, '%s' is rejected", e.FieldName, e.Message, e.RejectedValue)
}

// Validator définie les méthodes qu'implémentent les validateurs
type Validator[T any] interface {

	// IsValid dit si la donnée de type <T> fournié est valide
	IsValid() bool

	// IsValidOrPanic de même que IsValid, mais lance un panic
	IsValidOrPanic()

	// IsValidOrThrow de même que IsValid mais retourne une erreur
	IsValidOrThrow() error

	// GetName retourne le nom du validateur
	GetName() string

	// GetTarget retourne l'objet ayant fait l'objet de la validation
	GetTarget() T

	GetErrors() []ValidationError
}

type AbstractValidator[T any] struct {
	target     T
	name       string
	validators *validator.Validate
	ctx        context.Context
	errors     []ValidationError
	translator ut.Translator
}

func NewValidator[T any](target T, name string, c context.Context) *AbstractValidator[T] {
	valide := validator.New(
		validator.WithRequiredStructEnabled(),
		validator.WithPrivateFieldValidation(),
	)
	var englishTrans = english.New()
	uni := ut.New(englishTrans, englishTrans)
	trans, _ := uni.GetTranslator("en")
	_ = en_translation.RegisterDefaultTranslations(valide, trans)
	return &AbstractValidator[T]{
		target:     target,
		name:       name,
		validators: valide,
		ctx:        c,
		translator: trans,
	}
}

func (t *AbstractValidator[T]) GetName() string {
	return t.name
}

func (t *AbstractValidator[T]) GetTarget() T {
	return t.target
}

func (t *AbstractValidator[T]) IsValid() bool {
	return t.IsValidOrThrow() == nil
}
func (t *AbstractValidator[T]) IsValidOrPanic() {
	if e := t.IsValidOrThrow(); e != nil {
		panic(e)
	}
}

func (t *AbstractValidator[T]) IsValidOrThrow() error {
	var err error
	if t.ctx != nil {
		err = t.validators.StructCtx(t.ctx, t.target)
	} else {
		err = t.validators.Struct(t.target)
	}

	if err != nil {
		if _, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
			return ValidationErrors{Message: err.Error()}
		}

		if validateErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			var valerrors ValidationErrors
			for _, e := range validateErrs {
				valerrors.Errors = append(valerrors.Errors, ValidationError{
					Message:       e.Translate(t.translator),
					FieldName:     e.StructField(),
					RejectedValue: e.Value(),
				})
			}
			valerrors.Message = fmt.Sprintf("%s: %d validation error(s) found", t.name, len(validateErrs))
			t.errors = valerrors.Errors
			return valerrors
		}

		return err
	}

	return nil
}

func (t *AbstractValidator[T]) GetErrors() []ValidationError {
	return t.errors
}
