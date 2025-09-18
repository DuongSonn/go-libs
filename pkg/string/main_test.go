package _string

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Already snake case", "hello_world", "hello_world"},
		{"Camel case", "helloWorld", "hello_world"},
		{"Pascal case", "HelloWorld", "hello_world"},
		{"Mixed case", "HelloWORLD", "hello_world"},
		{"With hyphen", "hello-world", "hello_world"},
		{"With numbers", "hello123World", "hello123_world"},
		{"With consecutive uppercase", "HTTPResponse", "http_response"},
		{"With uppercase ID", "UserID", "user_id"},
		{"With uppercase and numbers", "User123ID", "user123_id"},
		{"With underscores already", "hello__world", "hello__world"},
		{"Single letter", "A", "a"},
		{"Complex mixed case", "ThisIsAnExampleString", "this_is_an_example_string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.input); got != tt.expected {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Already camel case", "helloWorld", "helloWorld"},
		{"Snake case", "hello_world", "helloWorld"},
		{"Pascal case", "HelloWorld", "helloWorld"},
		{"Mixed case", "HELLO_WORLD", "helloWorld"},
		{"With hyphen", "hello-world", "helloWorld"},
		{"With numbers", "hello_123_world", "hello123World"},
		{"With consecutive uppercase", "HTTP_response", "httpResponse"},
		{"With uppercase ID", "user_id", "userId"},
		{"With multiple underscores", "hello__world", "helloWorld"},
		{"Single letter", "a", "a"},
		{"Complex snake case", "this_is_an_example_string", "thisIsAnExampleString"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToCamelCase(tt.input); got != tt.expected {
				t.Errorf("ToCamelCase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Already pascal case", "HelloWorld", "HelloWorld"},
		{"Camel case", "helloWorld", "HelloWorld"},
		{"Snake case", "hello_world", "HelloWorld"},
		{"Mixed case", "HELLO_WORLD", "HelloWorld"},
		{"With hyphen", "hello-world", "HelloWorld"},
		{"With numbers", "hello_123_world", "Hello123World"},
		{"With consecutive uppercase", "HTTP_response", "HttpResponse"},
		{"With uppercase ID", "user_id", "UserId"},
		{"With multiple underscores", "hello__world", "HelloWorld"},
		{"Single letter", "a", "A"},
		{"Complex snake case", "this_is_an_example_string", "ThisIsAnExampleString"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToPascalCase(tt.input); got != tt.expected {
				t.Errorf("ToPascalCase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string", "", ""},
		{"Already kebab case", "hello-world", "hello-world"},
		{"Camel case", "helloWorld", "hello-world"},
		{"Pascal case", "HelloWorld", "hello-world"},
		{"Snake case", "hello_world", "hello-world"},
		{"Mixed case", "HELLO_WORLD", "hello-world"},
		{"With numbers", "hello123World", "hello123-world"},
		{"With consecutive uppercase", "HTTPResponse", "http-response"},
		{"With uppercase ID", "UserID", "user-id"},
		{"Single letter", "A", "a"},
		{"Complex mixed case", "ThisIsAnExampleString", "this-is-an-example-string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToKebabCase(tt.input); got != tt.expected {
				t.Errorf("ToKebabCase() = %v, want %v", got, tt.expected)
			}
		})
	}
}
