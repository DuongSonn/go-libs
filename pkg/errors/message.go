package _errors

import "fmt"

// ErrorMessage represents an error message with translations
type ErrorMessage struct {
	code         int               // Error code
	status       int               // HTTP status code
	translations map[string]string // Map of language -> message
}

// NewErrorMessage creates a new ErrorMessage
func NewErrorMessage(code int, status int) *ErrorMessage {
	return &ErrorMessage{
		code:         code,
		status:       status,
		translations: make(map[string]string),
	}
}

// WithTranslation adds a translation for the message
func (m *ErrorMessage) WithTranslation(lang, message string) *ErrorMessage {
	m.translations[lang] = message
	return m
}

// NewError creates an AppError from the ErrorMessage
func (m *ErrorMessage) NewError() *AppError {
	// Default to Vietnamese, fallback to English if not available
	message, ok := m.translations["vn"]
	if !ok {
		message, ok = m.translations["en"]
		if !ok {
			message = fmt.Sprintf("Error code: %d", m.code)
		}
	}

	return &AppError{
		Code:    m.code,
		Message: message,
		Status:  m.status,
	}
}

// NewErrorWithLang creates an AppError with a specific language
func (m *ErrorMessage) NewErrorWithLang(lang string) *AppError {
	message, ok := m.translations[lang]
	if !ok {
		// Fallback to Vietnamese
		message, ok = m.translations["vn"]
		if !ok {
			// Fallback to English
			message, ok = m.translations["en"]
			if !ok {
				message = fmt.Sprintf("Error code: %d", m.code)
			}
		}
	}

	return &AppError{
		Code:    m.code,
		Message: message,
		Status:  m.status,
	}
}

// NewErrorWithParams creates an AppError with parameters
func (m *ErrorMessage) NewErrorWithParams(params ...interface{}) *AppError {
	// Default to Vietnamese, fallback to English if not available
	message, ok := m.translations["vn"]
	if !ok {
		message, ok = m.translations["en"]
		if !ok {
			message = fmt.Sprintf("Error code: %d", m.code)
		}
	}

	return &AppError{
		Code:    m.code,
		Message: fmt.Sprintf(message, params...),
		Status:  m.status,
	}
}

// NewErrorWithLangAndParams creates an AppError with a specific language and parameters
func (m *ErrorMessage) NewErrorWithLangAndParams(lang string, params ...interface{}) *AppError {
	message, ok := m.translations[lang]
	if !ok {
		// Fallback to Vietnamese
		message, ok = m.translations["vn"]
		if !ok {
			// Fallback to English
			message, ok = m.translations["en"]
			if !ok {
				message = fmt.Sprintf("Error code: %d", m.code)
			}
		}
	}

	return &AppError{
		Code:    m.code,
		Message: fmt.Sprintf(message, params...),
		Status:  m.status,
	}
}

// ErrorRegistry is a registry for ErrorMessages
type ErrorRegistry struct {
	messages map[int]*ErrorMessage
}

// NewErrorRegistry creates a new ErrorRegistry
func NewErrorRegistry() *ErrorRegistry {
	return &ErrorRegistry{
		messages: make(map[int]*ErrorMessage),
	}
}

// Register registers an ErrorMessage
func (r *ErrorRegistry) Register(message *ErrorMessage) {
	// Check if message with same code already exists
	if _, ok := r.messages[message.code]; ok {
		fmt.Printf("Warning: Error message with code %d already exists\n", message.code)
	}

	r.messages[message.code] = message
}

// Get retrieves an ErrorMessage by code
func (r *ErrorRegistry) Get(code int) (*ErrorMessage, bool) {
	message, ok := r.messages[code]
	return message, ok
}

// MustGet retrieves an ErrorMessage by code, panics if not found
func (r *ErrorRegistry) MustGet(code int) *ErrorMessage {
	message, ok := r.messages[code]
	if !ok {
		panic(fmt.Sprintf("Error message with code %d not found", code))
	}
	return message
}

// Common error codes
const (
	ErrCodeUnknownError = 0
)

// DefaultRegistry is the default registry
var DefaultRegistry = NewErrorRegistry()

// Initialize default error messages
func init() {
	// Unknown error
	DefaultRegistry.Register(
		NewErrorMessage(ErrCodeUnknownError, 500).
			WithTranslation("vn", "Lỗi không xác định").
			WithTranslation("en", "Unknown error"),
	)
}
