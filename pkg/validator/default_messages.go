package _validator

import (
	_errors "go-libs/pkg/errors"
)

// Validation error codes
const (
	// Base error code
	ValidationErrorBase = 10000 // Start from 10000 to avoid conflicts with other error codes

	// Specific error codes
	ValidationErrorRequired     = ValidationErrorBase + 1
	ValidationErrorMin          = ValidationErrorBase + 2
	ValidationErrorMax          = ValidationErrorBase + 3
	ValidationErrorEmail        = ValidationErrorBase + 4
	ValidationErrorURL          = ValidationErrorBase + 5
	ValidationErrorAlpha        = ValidationErrorBase + 6
	ValidationErrorNumeric      = ValidationErrorBase + 7
	ValidationErrorAlphaNumeric = ValidationErrorBase + 8
	ValidationErrorLen          = ValidationErrorBase + 9
	ValidationErrorEq           = ValidationErrorBase + 10
	ValidationErrorNe           = ValidationErrorBase + 11
	ValidationErrorGt           = ValidationErrorBase + 12
	ValidationErrorGte          = ValidationErrorBase + 13
	ValidationErrorLt           = ValidationErrorBase + 14
	ValidationErrorLte          = ValidationErrorBase + 15
	ValidationErrorOneOf        = ValidationErrorBase + 16
	ValidationErrorUnique       = ValidationErrorBase + 17
)

// registerDefaultValidationErrors registers default validation error messages
func registerDefaultValidationErrors(registry *_errors.ErrorRegistry) {
	registry.Register(_errors.NewErrorMessage(ValidationErrorBase, 400).
		WithTranslation(_errors.LangVN, "Tham số thiếu hoặc không hợp lệ. Vui lòng kiểm tra lại!").
		WithTranslation(_errors.LangEN, "Missing or invalid parameters. Please check again!"))

	// Register required error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorRequired, 400).
			WithTranslation(_errors.LangVN, "{field} là bắt buộc").
			WithTranslation(_errors.LangEN, "{field} is required"),
	)

	// Register min error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorMin, 400).
			WithTranslation(_errors.LangVN, "{field} phải có ít nhất {param}").
			WithTranslation(_errors.LangEN, "{field} must be at least {param}"),
	)

	// Register max error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorMax, 400).
			WithTranslation(_errors.LangVN, "{field} không được vượt quá {param}").
			WithTranslation(_errors.LangEN, "{field} must be at most {param}"),
	)

	// Register email error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorEmail, 400).
			WithTranslation(_errors.LangVN, "{field} phải là địa chỉ email hợp lệ").
			WithTranslation(_errors.LangEN, "{field} must be a valid email address"),
	)

	// Register url error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorURL, 400).
			WithTranslation(_errors.LangVN, "{field} phải là URL hợp lệ").
			WithTranslation(_errors.LangEN, "{field} must be a valid URL"),
	)

	// Register alpha error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorAlpha, 400).
			WithTranslation(_errors.LangVN, "{field} chỉ được chứa chữ cái").
			WithTranslation(_errors.LangEN, "{field} must contain only letters"),
	)

	// Register numeric error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorNumeric, 400).
			WithTranslation(_errors.LangVN, "{field} chỉ được chứa số").
			WithTranslation(_errors.LangEN, "{field} must contain only numbers"),
	)

	// Register alphanum error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorAlphaNumeric, 400).
			WithTranslation(_errors.LangVN, "{field} chỉ được chứa chữ cái và số").
			WithTranslation(_errors.LangEN, "{field} must contain only letters and numbers"),
	)

	// Register len error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorLen, 400).
			WithTranslation(_errors.LangVN, "{field} phải có đúng {param} ký tự").
			WithTranslation(_errors.LangEN, "{field} must be exactly {param} characters long"),
	)

	// Register eq error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorEq, 400).
			WithTranslation(_errors.LangVN, "{field} phải bằng {param}").
			WithTranslation(_errors.LangEN, "{field} must be equal to {param}"),
	)

	// Register ne error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorNe, 400).
			WithTranslation(_errors.LangVN, "{field} không được bằng {param}").
			WithTranslation(_errors.LangEN, "{field} must not be equal to {param}"),
	)

	// Register gt error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorGt, 400).
			WithTranslation(_errors.LangVN, "{field} phải lớn hơn {param}").
			WithTranslation(_errors.LangEN, "{field} must be greater than {param}"),
	)

	// Register gte error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorGte, 400).
			WithTranslation(_errors.LangVN, "{field} phải lớn hơn hoặc bằng {param}").
			WithTranslation(_errors.LangEN, "{field} must be greater than or equal to {param}"),
	)

	// Register lt error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorLt, 400).
			WithTranslation(_errors.LangVN, "{field} phải nhỏ hơn {param}").
			WithTranslation(_errors.LangEN, "{field} must be less than {param}"),
	)

	// Register lte error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorLte, 400).
			WithTranslation(_errors.LangVN, "{field} phải nhỏ hơn hoặc bằng {param}").
			WithTranslation(_errors.LangEN, "{field} must be less than or equal to {param}"),
	)

	// Register oneof error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorOneOf, 400).
			WithTranslation(_errors.LangVN, "{field} phải là một trong [{param}]").
			WithTranslation(_errors.LangEN, "{field} must be one of [{param}]"),
	)

	// Register unique error message
	registry.Register(
		_errors.NewErrorMessage(ValidationErrorUnique, 400).
			WithTranslation(_errors.LangVN, "{field} phải chứa các giá trị duy nhất").
			WithTranslation(_errors.LangEN, "{field} must contain unique values"),
	)
}
