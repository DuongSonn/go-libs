package _validator

import (
	"reflect"
	"strings"

	_errors "go-libs/pkg/errors"

	"github.com/go-playground/validator/v10"
)

// ErrorValidator wraps go-playground/validator with errors package
type ErrorValidator struct {
	validate *validator.Validate
	errReg   *_errors.ErrorRegistry
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Param   string `json:"param,omitempty"`
	Value   any    `json:"value,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("validation failed: ")
	for i, err := range e {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Field)
		sb.WriteString(": ")
		sb.WriteString(err.Message)
	}
	return sb.String()
}

// NewValidator creates a new validator with error registry
func NewValidator(errReg *_errors.ErrorRegistry) *ErrorValidator {
	// Create validator instance
	validate := validator.New()

	// Register function to get JSON tag as field name
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	// Initialize validator with error registry
	v := &ErrorValidator{
		validate: validate,
		errReg:   errReg,
	}

	// Register default validation errors
	registerDefaultValidationErrors(errReg)

	return v
}

// RegisterCustomValidator registers a custom validation function
func (v *ErrorValidator) RegisterCustomValidator(tag string, customError *_errors.ErrorMessage, fn validator.Func) error {
	// Register validation function
	if err := v.validate.RegisterValidation(tag, fn); err != nil {
		return err
	}

	// Register custom error message, if not the package will use the default error message
	v.errReg.Register(customError)

	return nil
}

// Validate validates a struct and returns validation errors
func (v *ErrorValidator) Validate(value interface{}, lang string) (ValidationErrors, error) {
	// Validate the struct
	err := v.validate.Struct(value)
	if err == nil {
		return nil, nil
	}

	// Convert validation errors
	validationErrors := make(ValidationErrors, 0)

	validatorErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, err
	}

	for _, e := range validatorErrs {
		// Get validation error code for the tag
		errorCode := v.getValidationErrorCode(e.Tag())

		var message string
		var appError *_errors.AppError

		// Try to find error message in registry
		if errMsg, ok := v.errReg.Get(errorCode); ok {
			// Use message from error registry
			appError = errMsg.NewErrorWithLangAndParams(lang, e.Field(), e.Param())
			message = appError.Message
		} else {
			// Use default message
			message = v.getDefaultMessage(lang)
		}

		validationError := ValidationError{
			Field:   e.Field(),
			Message: message,
			Tag:     e.Tag(),
			Param:   e.Param(),
			Value:   e.Value(),
			Code:    errorCode,
		}

		validationErrors = append(validationErrors, validationError)
	}

	return validationErrors, nil
}

// ValidateVar validates a single variable
func (v *ErrorValidator) ValidateVar(value interface{}, tag string, fieldName string, lang string) (ValidationErrors, error) {
	// Validate the variable
	err := v.validate.Var(value, tag)
	if err == nil {
		return nil, nil
	}

	// Convert validation errors
	validationErrors := make(ValidationErrors, 0)

	validatorErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, err
	}

	for _, e := range validatorErrs {
		// Get validation error code for the tag
		errorCode := v.getValidationErrorCode(e.Tag())

		var message string
		var appError *_errors.AppError

		// Try to find error message in registry
		if errMsg, ok := v.errReg.Get(errorCode); ok {
			// Use message from error registry
			appError = errMsg.NewErrorWithLangAndParams(lang, fieldName, e.Param())
			message = appError.Message
		} else {
			// Use default message
			message = v.getDefaultMessage(lang)
		}

		validationError := ValidationError{
			Field:   fieldName,
			Message: message,
			Tag:     e.Tag(),
			Param:   e.Param(),
			Value:   e.Value(),
			Code:    errorCode,
		}

		validationErrors = append(validationErrors, validationError)
	}

	return validationErrors, nil
}

// ValidateMap validates a map of field-value pairs
func (v *ErrorValidator) ValidateMap(values map[string]interface{}, rules map[string]string, lang string) (ValidationErrors, error) {
	// Validate each field
	validationErrors := make(ValidationErrors, 0)

	for field, rule := range rules {
		value, exists := values[field]
		if !exists {
			// Skip validation if field doesn't exist
			continue
		}

		// Validate the field
		err := v.validate.Var(value, rule)
		if err == nil {
			continue
		}

		// Convert validation errors
		validatorErrs, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, err
		}

		for _, e := range validatorErrs {
			// Get validation error code for the tag
			errorCode := v.getValidationErrorCode(e.Tag())

			var message string
			var appError *_errors.AppError

			// Try to find error message in registry
			if errMsg, ok := v.errReg.Get(errorCode); ok {
				// Use message from error registry
				appError = errMsg.NewErrorWithLangAndParams(lang, field, e.Param())
				message = appError.Message
			} else {
				// Use default message
				message = v.getDefaultMessage(lang)
			}

			validationError := ValidationError{
				Field:   field,
				Message: message,
				Tag:     e.Tag(),
				Param:   e.Param(),
				Value:   e.Value(),
				Code:    errorCode,
			}

			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors, nil
}

// GetValidator returns the underlying validator instance
func (v *ErrorValidator) GetValidator() *validator.Validate {
	return v.validate
}

// getDefaultMessage returns a default error message for the given language
func (v *ErrorValidator) getDefaultMessage(lang string) string {
	return v.errReg.MustGet(_errors.ErrCodeUnknownError).NewErrorWithLang(lang).Message
}

// getValidationErrorCode returns the error code for a validation tag
func (v *ErrorValidator) getValidationErrorCode(tag string) int {
	switch tag {
	case "required":
		return ValidationErrorRequired
	case "min":
		return ValidationErrorMin
	case "max":
		return ValidationErrorMax
	case "email":
		return ValidationErrorEmail
	case "url":
		return ValidationErrorURL
	case "alpha":
		return ValidationErrorAlpha
	case "numeric":
		return ValidationErrorNumeric
	case "alphanum":
		return ValidationErrorAlphaNumeric
	case "len":
		return ValidationErrorLen
	case "eq":
		return ValidationErrorEq
	case "ne":
		return ValidationErrorNe
	case "gt":
		return ValidationErrorGt
	case "gte":
		return ValidationErrorGte
	case "lt":
		return ValidationErrorLt
	case "lte":
		return ValidationErrorLte
	case "oneof":
		return ValidationErrorOneOf
	case "unique":
		return ValidationErrorUnique
	default:
		return ValidationErrorBase
	}
}
