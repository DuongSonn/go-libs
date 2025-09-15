# Error Handling Package

A multilingual error handling package with registry pattern for Go applications.

## Features

-   Registry pattern for module-specific error codes
-   Multilingual support (Vietnamese, English)
-   Parameterized error messages
-   HTTP status code integration
-   Type-safe error handling

## Basic Usage

```go
// Create a registry
registry := errors.NewErrorRegistry()

// Register error messages
registry.Register(
    errors.NewErrorMessage(1000, http.StatusBadRequest).
        WithTranslation(errors.LangEN, "Invalid request").
        WithTranslation(errors.LangVN, "Yêu cầu không hợp lệ"),
)

// Create an error
err := registry.MustGet(1000).NewError()

// Create an error with specific language
errEN := registry.MustGet(1000).NewErrorWithLang(errors.LangEN)

// Create an error with parameters
errWithParams := registry.MustGet(1000).NewErrorWithParams("param1", 123)
```

## Cash Application Example

```go
// Define error codes
const (
    // Transaction errors
    ErrInsufficientFunds = 1000
    ErrExceedsDailyLimit = 1001

    // Account errors
    ErrAccountNotFound = 2000
    ErrAccountLocked   = 2001
)

// Create registry for cash module
cashErrors := errors.NewErrorRegistry()

// Register errors
cashErrors.Register(
    errors.NewErrorMessage(ErrInsufficientFunds, http.StatusBadRequest).
        WithTranslation(errors.LangEN, "Insufficient funds").
        WithTranslation(errors.LangVN, "Số dư không đủ"),
)

cashErrors.Register(
    errors.NewErrorMessage(ErrAccountNotFound, http.StatusNotFound).
        WithTranslation(errors.LangEN, "Account not found").
        WithTranslation(errors.LangVN, "Không tìm thấy tài khoản"),
)

// Register parameterized error
cashErrors.Register(
    errors.NewErrorMessage(ErrExceedsDailyLimit, http.StatusBadRequest).
        WithTranslation(errors.LangEN, "Transaction exceeds daily limit of %s").
        WithTranslation(errors.LangVN, "Giao dịch vượt quá hạn mức ngày %s"),
)

// Use in functions
func Withdraw(accountID string, amount float64) error {
    // Check if account exists
    if !accountExists(accountID) {
        return cashErrors.MustGet(ErrAccountNotFound).NewError()
    }

    // Check if sufficient funds
    if getBalance(accountID) < amount {
        return cashErrors.MustGet(ErrInsufficientFunds).NewError()
    }

    // Check daily limit
    if exceedsDailyLimit(amount) {
        return cashErrors.MustGet(ErrExceedsDailyLimit).
            NewErrorWithParams("$1,000.00")
    }

    // Process withdrawal
    return nil
}

// Handle errors
func HandleWithdrawal(w http.ResponseWriter, r *http.Request) {
    // Process request...
    err := Withdraw("123", 500.0)

    if err != nil {
        if appErr, ok := err.(*errors.AppError); ok {
            // Return JSON with error details
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(appErr.GetStatus())
            json.NewEncoder(w).Encode(appErr)
            return
        }

        // Handle unexpected errors
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    // Success response...
}
```

## Error Handling

```go
err := someFunction()
if err != nil {
    if appErr, ok := err.(*errors.AppError); ok {
        // Handle by error code
        switch appErr.GetCode() {
        case ErrInsufficientFunds:
            // Handle insufficient funds
        case ErrAccountNotFound:
            // Handle account not found
        default:
            // Handle other errors
        }

        // Or handle by HTTP status
        switch appErr.GetStatus() {
        case http.StatusBadRequest:
            // Handle bad request errors
        case http.StatusNotFound:
            // Handle not found errors
        }
    }
}
```
