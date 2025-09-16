package _time

import (
	"testing"
	"time"
)

func TestGetStartOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Regular date",
			input:    time.Date(2023, 5, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Already at start of day",
			input:    time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "End of day",
			input:    time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.UTC),
			expected: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "With timezone",
			input:    time.Date(2023, 5, 15, 14, 30, 45, 123456789, time.FixedZone("UTC+7", 7*60*60)),
			expected: time.Date(2023, 5, 15, 0, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStartOfDay(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("GetStartOfDay() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEndOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "Regular date",
			input:    time.Date(2023, 5, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "Already at start of day",
			input:    time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
			expected: time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "Already at end of day",
			input:    time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.UTC),
			expected: time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "With timezone",
			input:    time.Date(2023, 5, 15, 14, 30, 45, 123456789, time.FixedZone("UTC+7", 7*60*60)),
			expected: time.Date(2023, 5, 15, 23, 59, 59, 999999999, time.FixedZone("UTC+7", 7*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEndOfDay(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("GetEndOfDay() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetStartOfQuarter(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "First month of Q1 (January)",
			input:    time.Date(2023, 1, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Second month of Q1 (February)",
			input:    time.Date(2023, 2, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "Third month of Q1 (March)",
			input:    time.Date(2023, 3, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "First month of Q2 (April)",
			input:    time.Date(2023, 4, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "First month of Q3 (July)",
			input:    time.Date(2023, 7, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "First month of Q4 (October)",
			input:    time.Date(2023, 10, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "With timezone",
			input:    time.Date(2023, 5, 15, 14, 30, 45, 123456789, time.FixedZone("UTC+7", 7*60*60)),
			expected: time.Date(2023, 4, 1, 0, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStartOfQuarter(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("GetStartOfQuarter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEndOfQuarter(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			name:     "First month of Q1 (January)",
			input:    time.Date(2023, 1, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 3, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "Second month of Q1 (February)",
			input:    time.Date(2023, 2, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 3, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "First month of Q2 (April)",
			input:    time.Date(2023, 4, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 6, 30, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "First month of Q3 (July)",
			input:    time.Date(2023, 7, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 9, 30, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "First month of Q4 (October)",
			input:    time.Date(2023, 10, 15, 14, 30, 45, 123456789, time.UTC),
			expected: time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.UTC),
		},
		{
			name:     "With timezone",
			input:    time.Date(2023, 5, 15, 14, 30, 45, 123456789, time.FixedZone("UTC+7", 7*60*60)),
			expected: time.Date(2023, 6, 30, 23, 59, 59, 999999999, time.FixedZone("UTC+7", 7*60*60)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEndOfQuarter(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("GetEndOfQuarter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected bool
	}{
		{
			name:     "Monday",
			input:    time.Date(2023, 5, 15, 12, 0, 0, 0, time.UTC), // Monday
			expected: false,
		},
		{
			name:     "Tuesday",
			input:    time.Date(2023, 5, 16, 12, 0, 0, 0, time.UTC), // Tuesday
			expected: false,
		},
		{
			name:     "Wednesday",
			input:    time.Date(2023, 5, 17, 12, 0, 0, 0, time.UTC), // Wednesday
			expected: false,
		},
		{
			name:     "Thursday",
			input:    time.Date(2023, 5, 18, 12, 0, 0, 0, time.UTC), // Thursday
			expected: false,
		},
		{
			name:     "Friday",
			input:    time.Date(2023, 5, 19, 12, 0, 0, 0, time.UTC), // Friday
			expected: false,
		},
		{
			name:     "Saturday",
			input:    time.Date(2023, 5, 20, 12, 0, 0, 0, time.UTC), // Saturday
			expected: true,
		},
		{
			name:     "Sunday",
			input:    time.Date(2023, 5, 21, 12, 0, 0, 0, time.UTC), // Sunday
			expected: true,
		},
		{
			name:     "With timezone",
			input:    time.Date(2023, 5, 20, 12, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60)), // Saturday
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWeekend(tt.input)
			if result != tt.expected {
				t.Errorf("IsWeekend() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsLeapYear(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		expected bool
	}{
		{
			name:     "Regular leap year (divisible by 4)",
			year:     2020,
			expected: true,
		},
		{
			name:     "Regular non-leap year",
			year:     2021,
			expected: false,
		},
		{
			name:     "Century year (divisible by 100 but not by 400)",
			year:     1900,
			expected: false,
		},
		{
			name:     "Century leap year (divisible by 400)",
			year:     2000,
			expected: true,
		},
		{
			name:     "Future leap year",
			year:     2024,
			expected: true,
		},
		{
			name:     "Future non-leap year",
			year:     2025,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLeapYear(tt.year)
			if result != tt.expected {
				t.Errorf("IsLeapYear() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDaysInMonth(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		month    time.Month
		expected int
	}{
		{
			name:     "January (31 days)",
			year:     2023,
			month:    time.January,
			expected: 31,
		},
		{
			name:     "February non-leap year (28 days)",
			year:     2023,
			month:    time.February,
			expected: 28,
		},
		{
			name:     "February leap year (29 days)",
			year:     2024,
			month:    time.February,
			expected: 29,
		},
		{
			name:     "April (30 days)",
			year:     2023,
			month:    time.April,
			expected: 30,
		},
		{
			name:     "Century non-leap year February",
			year:     1900,
			month:    time.February,
			expected: 28,
		},
		{
			name:     "Century leap year February",
			year:     2000,
			month:    time.February,
			expected: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DaysInMonth(tt.year, tt.month)
			if result != tt.expected {
				t.Errorf("DaysInMonth() = %v, want %v", result, tt.expected)
			}
		})
	}
}
