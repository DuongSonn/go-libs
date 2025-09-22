# Validator Package

A powerful validation package for Go applications that integrates with the error registry system to provide multilingual validation error messages.

## Features

-   Built on top of [go-playground/validator](https://github.com/go-playground/validator)
-   Multilingual error messages (English and Vietnamese)
-   Integration with the error registry system
-   Support for struct, variable, and map validation
-   Custom validator registration
-   Numeric error codes for validation errors

## Installation

```bash
go get -u go-libs/pkg/validator
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "go-libs/pkg/errors"
    "go-libs/pkg/validator"
)

// User represents a user model
type User struct {
    ID       int    `json:"id" validate:"required,min=1"`
    Username string `json:"username" validate:"required,min=3,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,min=18"`
}

func main() {
    // Create a new error registry
    errReg := errors.NewErrorRegistry()

    // Create a new validator with the error registry
    v := validator._validator.NewValidator(errReg)

    // Create a user with invalid data
    user := User{
        ID:       0,
        Username: "jo",
        Email:    "invalid-email",
        Age:      16,
    }

    // Validate the user with English locale
    errors, err := v.Validate(user, "en")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    if len(errors) > 0 {
        fmt.Printf("Found %d validation errors:\n", len(errors))
        for _, e := range errors {
            fmt.Printf("- %s: %s (tag: %s, code: %d)\n", e.Field, e.Message, e.Tag, e.Code)
        }
    }
}
```

### Custom Validator

```go
package main

import (
    "fmt"
    "regexp"
    "go-libs/pkg/errors"
    "go-libs/pkg/validator"
    "github.com/go-playground/validator/v10"
)

// Define a custom validation function for Vietnamese phone numbers
func validateVNPhone(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    // Regex for Vietnamese phone numbers
    // - Starts with 0 or +84
    // - Followed by 9 or 3, 5, 7, 8, 9 (depending on carrier)
    // - Followed by 8 digits
    vnPhoneRegex := regexp.MustCompile(`^(0|\+84)([3|5|7|8|9])([0-9]{8})$`)
    return vnPhoneRegex.MatchString(phone)
}

func main() {
    // Create a new error registry
    errReg := errors.NewErrorRegistry()

    // Create a new validator with the error registry
    v := validator._validator.NewValidator(errReg)

    // Create custom error message for Vietnamese phone validation
    vnPhoneError := errors.NewErrorMessage(10100, 400).
        WithTranslation("vn", "{field} phải là số điện thoại Việt Nam hợp lệ").
        WithTranslation("en", "{field} must be a valid Vietnamese phone number")

    // Register custom validator with tag "vnphone"
    v.RegisterCustomValidator("vnphone", vnPhoneError, validateVNPhone)

    // Use the custom validator
    type Contact struct {
        Name  string `json:"name" validate:"required"`
        Phone string `json:"phone" validate:"required,vnphone"`
    }

    contact := Contact{
        Name:  "Nguyễn Văn A",
        Phone: "invalid-phone",
    }

    errors, _ := v.Validate(contact, "vn")
    if len(errors) > 0 {
        for _, e := range errors {
            fmt.Printf("Lỗi: %s - %s\n", e.Field, e.Message)
        }
    }
}
```

### Validating Single Variables

```go
// Validate a single variable
email := "invalid-email"
errors, _ := v.ValidateVar(email, "required,email", "Email", "en")
if len(errors) > 0 {
    fmt.Printf("Email validation error: %s\n", errors[0].Message)
}
```

### Validating Maps

```go
// Validate a map of values
values := map[string]interface{}{
    "name":  "Jo",
    "email": "invalid-email",
    "age":   16,
}

rules := map[string]string{
    "name":  "required,min=3",
    "email": "required,email",
    "age":   "required,min=18",
}

errors, _ := v.ValidateMap(values, rules, "en")
if len(errors) > 0 {
    for _, e := range errors {
        fmt.Printf("Map validation error: %s - %s\n", e.Field, e.Message)
    }
}
```

## API Reference

### Types

#### `ErrorValidator`

```go
type ErrorValidator struct {
    // Contains unexported fields
}
```

#### `ValidationError`

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Tag     string `json:"tag"`
    Param   string `json:"param,omitempty"`
    Value   any    `json:"value,omitempty"`
    Code    int    `json:"code,omitempty"`
}
```

#### `ValidationErrors`

```go
type ValidationErrors []ValidationError
```

### Functions

#### `NewValidator`

```go
func NewValidator(errReg *_errors.ErrorRegistry) *ErrorValidator
```

Creates a new validator with the provided error registry.

#### `RegisterCustomValidator`

```go
func (v *ErrorValidator) RegisterCustomValidator(tag string, customError *_errors.ErrorMessage, fn validator.Func) error
```

Registers a custom validation function with the specified tag and error message.

#### `Validate`

```go
func (v *ErrorValidator) Validate(value interface{}, lang string) (ValidationErrors, error)
```

Validates a struct and returns validation errors.

#### `ValidateVar`

```go
func (v *ErrorValidator) ValidateVar(value interface{}, tag string, fieldName string, lang string) (ValidationErrors, error)
```

Validates a single variable with the specified tag.

#### `ValidateMap`

```go
func (v *ErrorValidator) ValidateMap(values map[string]interface{}, rules map[string]string, lang string) (ValidationErrors, error)
```

Validates a map of field-value pairs with the specified rules.

#### `GetValidator`

```go
func (v *ErrorValidator) GetValidator() *validator.Validate
```

Returns the underlying validator instance.

## License

This package is part of the go-libs project.
