package _errors

// AppError represents an application error with code, message and HTTP status
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// GetCode returns the error code
func (e *AppError) GetCode() int {
	return e.Code
}

// GetStatus returns the HTTP status code
func (e *AppError) GetStatus() int {
	return e.Status
}

// WithMessage creates a copy of the error with a new message
func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: message,
		Status:  e.Status,
	}
}
