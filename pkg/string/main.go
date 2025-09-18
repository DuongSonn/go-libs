package _string

import (
	"regexp"
	"strings"
)

// ToSnakeCase converts a string to snake_case format.
// It handles camelCase, PascalCase, and kebab-case inputs.
// Examples:
//   - "HelloWorld" -> "hello_world"
//   - "helloWorld" -> "hello_world"
//   - "HelloWORLD" -> "hello_world"
//   - "hello-world" -> "hello_world"
//   - "hello_world" -> "hello_world" (unchanged)
//   - "HTTP_Response" -> "http_response"
//   - "UserID" -> "user_id"
func ToSnakeCase(str string) string {
	// Handle special cases
	if len(str) == 0 {
		return str
	}

	// Replace hyphens with underscores
	str = strings.ReplaceAll(str, "-", "_")

	// Use regex to insert underscores between:
	// 1. A lowercase letter followed by an uppercase letter
	// 2. A number followed by a letter
	// 3. A letter followed by a number
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	str = re.ReplaceAllString(str, "${1}_${2}")

	// Handle consecutive uppercase letters properly
	// e.g., "HTTPResponse" -> "http_response"
	re = regexp.MustCompile(`([A-Z])([A-Z][a-z])`)
	str = re.ReplaceAllString(str, "${1}_${2}")

	// Convert to lowercase
	return strings.ToLower(str)
}

// ToCamelCase converts a string to camelCase format.
// It handles snake_case, PascalCase, and kebab-case inputs.
// Examples:
//   - "hello_world" -> "helloWorld"
//   - "HelloWorld" -> "helloWorld"
//   - "hello-world" -> "helloWorld"
//   - "HELLO_WORLD" -> "helloWorld"
//   - "HTTP_response" -> "httpResponse"
//   - "user_id" -> "userId"
//   - "helloWorld" -> "helloWorld" (unchanged)
func ToCamelCase(str string) string {
	// First convert to snake case to normalize
	str = ToSnakeCase(str)

	// Split by underscore
	words := strings.Split(str, "_")

	// First word stays lowercase, capitalize the rest
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
		}
	}

	return strings.Join(words, "")
}

// ToPascalCase converts a string to PascalCase format.
// It handles snake_case, camelCase, and kebab-case inputs.
// Examples:
//   - "hello_world" -> "HelloWorld"
//   - "helloWorld" -> "HelloWorld"
//   - "hello-world" -> "HelloWorld"
//   - "HELLO_WORLD" -> "HelloWorld"
//   - "HTTP_response" -> "HttpResponse"
//   - "user_id" -> "UserId"
//   - "HelloWorld" -> "HelloWorld" (unchanged)
func ToPascalCase(str string) string {
	// First convert to camel case
	str = ToCamelCase(str)

	// Capitalize the first letter
	if len(str) > 0 {
		str = strings.ToUpper(str[:1]) + str[1:]
	}

	return str
}

// Examples:
//   - "HelloWorld" -> "hello-world"
//   - "helloWorld" -> "hello-world"
//   - "hello_world" -> "hello-world"
//   - "HELLO_WORLD" -> "hello-world"
func ToKebabCase(str string) string {
	// First convert to snake case
	snake := ToSnakeCase(str)

	// Replace underscores with hyphens
	return strings.ReplaceAll(snake, "_", "-")
}
