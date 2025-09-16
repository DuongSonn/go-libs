package _time

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetStartOfDay returns the start time of the day (00:00:00)
func GetStartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// GetEndOfDay returns the end time of the day (23:59:59.999999999)
func GetEndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-1), t.Location())
}

// GetStartOfMonth returns the first day of the month (day 1, 00:00:00)
func GetStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfMonth returns the last day of the month (23:59:59.999999999)
func GetEndOfMonth(t time.Time) time.Time {
	// Get the first day of the next month, then subtract 1 nanosecond
	return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location()).Add(-time.Nanosecond)
}

// GetStartOfYear returns the first day of the year (01/01, 00:00:00)
func GetStartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfYear returns the last day of the year (12/31, 23:59:59.999999999)
func GetEndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), time.December, 31, 23, 59, 59, int(time.Second-1), t.Location())
}

// CountDaysBetween calculates the number of days between two time points
// The result is always positive, regardless of the order of start and end
func CountDaysBetween(start, end time.Time) int {
	// Normalize times to start of day to ensure accuracy
	startDay := GetStartOfDay(start)
	endDay := GetStartOfDay(end)

	// Ensure end >= start
	if endDay.Before(startDay) {
		startDay, endDay = endDay, startDay
	}

	// Calculate days
	return int(math.Round(endDay.Sub(startDay).Hours() / 24))
}

// CountMonthsBetween calculates the number of months between two time points
// The result is always positive, regardless of the order of start and end
func CountMonthsBetween(start, end time.Time) int {
	// Ensure end >= start
	if end.Before(start) {
		start, end = end, start
	}

	// Calculate months
	months := (end.Year()-start.Year())*12 + int(end.Month()) - int(start.Month())

	// Adjust if the day of month of end is less than start
	if end.Day() < start.Day() {
		months--
	}

	return months
}

// CountYearsBetween calculates the number of years between two time points
// The result is always positive, regardless of the order of start and end
func CountYearsBetween(start, end time.Time) int {
	// Ensure end >= start
	if end.Before(start) {
		start, end = end, start
	}

	// Calculate years
	years := end.Year() - start.Year()

	// Adjust if the month of end is less than start, or months are equal but day is less
	if end.Month() < start.Month() || (end.Month() == start.Month() && end.Day() < start.Day()) {
		years--
	}

	return years
}

// GetStartOfWeek returns the first day of the week (Monday, 00:00:00)
func GetStartOfWeek(t time.Time) time.Time {
	// Get the number of days to subtract to reach the beginning of the week (Monday)
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday (0) needs to subtract 6 days to reach the previous Monday
		weekday = 7
	}
	daysToSubtract := weekday - 1 // Subtract to reach Monday (1)

	// Subtract days and set to 00:00:00
	return GetStartOfDay(t.AddDate(0, 0, -daysToSubtract))
}

// GetEndOfWeek returns the last day of the week (Sunday, 23:59:59.999999999)
func GetEndOfWeek(t time.Time) time.Time {
	// Get the start of the week, then add 6 days to reach Sunday
	startOfWeek := GetStartOfWeek(t)
	return GetEndOfDay(startOfWeek.AddDate(0, 0, 6))
}

// GetStartOfQuarter returns the first day of the quarter (00:00:00)
func GetStartOfQuarter(t time.Time) time.Time {
	quarter := (int(t.Month()) - 1) / 3
	firstMonthOfQuarter := time.Month(quarter*3 + 1)
	return time.Date(t.Year(), firstMonthOfQuarter, 1, 0, 0, 0, 0, t.Location())
}

// GetEndOfQuarter returns the last day of the quarter (23:59:59.999999999)
func GetEndOfQuarter(t time.Time) time.Time {
	quarter := (int(t.Month()) - 1) / 3
	lastMonthOfQuarter := time.Month(quarter*3 + 3)
	return GetEndOfMonth(time.Date(t.Year(), lastMonthOfQuarter, 1, 0, 0, 0, 0, t.Location()))
}

// AddBusinessDays adds the specified number of business days (excluding Sat, Sun) to the time
func AddBusinessDays(t time.Time, days int) time.Time {
	// If days = 0, return the original time
	if days == 0 {
		return t
	}

	result := t
	step := 1
	if days < 0 {
		step = -1
		days = -days
	}

	for i := 0; i < days; {
		result = result.AddDate(0, 0, step)
		// If not a weekend (Sat, Sun), increment the counter
		if result.Weekday() != time.Saturday && result.Weekday() != time.Sunday {
			i++
		}
	}

	return result
}

// IsWeekend checks if a day is a weekend (Saturday or Sunday)
func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

// IsLeapYear checks if a year is a leap year
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in the specified month
func DaysInMonth(year int, month time.Month) int {
	// The 0th day of the next month is the last day of the current month
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// FormatDuration formats a duration into a human-readable string
// Examples: "2h 30m", "3d 4h", "1y 2mo"
func FormatDuration(d time.Duration) string {
	// If duration is less than a minute, use Go's default format
	if d < time.Minute {
		return d.String()
	}

	var result string

	// Calculate time units
	days := int(d.Hours()) / 24
	years := days / 365
	days = days % 365
	months := days / 30
	days = days % 30
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	// Build the result string
	if years > 0 {
		result += fmt.Sprintf("%dy ", years)
	}
	if months > 0 {
		result += fmt.Sprintf("%dmo ", months)
	}
	if days > 0 {
		result += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		result += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		result += fmt.Sprintf("%dm", minutes)
	}

	return strings.TrimSpace(result)
}

// ParseDuration parses a human-readable duration string into time.Duration
// Supports standard Go duration units (ns, us, ms, s, m, h) plus:
// - d: days (24 hours)
// - w: weeks (7 days)
// - mo: months (30 days)
// - y: years (365 days)
// Examples: "2h30m", "3d4h", "1y2mo", "2w3d12h"
func ParseDuration(durationStr string) (time.Duration, error) {
	// First try standard Go duration parsing
	if d, err := time.ParseDuration(durationStr); err == nil {
		return d, nil
	}

	// Define extended units
	day := 24 * time.Hour
	week := 7 * day
	month := 30 * day
	year := 365 * day

	// Clean the input string
	s := strings.TrimSpace(durationStr)
	s = strings.ToLower(s)

	// Define regex patterns for each unit
	patterns := []struct {
		regex *regexp.Regexp
		unit  time.Duration
	}{
		{regexp.MustCompile(`(\d+)y`), year},
		{regexp.MustCompile(`(\d+)mo`), month},
		{regexp.MustCompile(`(\d+)w`), week},
		{regexp.MustCompile(`(\d+)d`), day},
		{regexp.MustCompile(`(\d+)h`), time.Hour},
		{regexp.MustCompile(`(\d+)m([^o]|$)`), time.Minute}, // m not followed by 'o' or at the end of string
		{regexp.MustCompile(`(\d+)s`), time.Second},
		{regexp.MustCompile(`(\d+)ms`), time.Millisecond},
		{regexp.MustCompile(`(\d+)us`), time.Microsecond},
		{regexp.MustCompile(`(\d+)ns`), time.Nanosecond},
	}

	var duration time.Duration
	foundMatch := false

	// Extract and sum all time components
	for _, pattern := range patterns {
		matches := pattern.regex.FindAllStringSubmatch(s, -1)
		for _, match := range matches {
			if len(match) > 1 {
				value, err := strconv.Atoi(match[1])
				if err != nil {
					return 0, fmt.Errorf("invalid value in duration: %s", match[0])
				}
				duration += time.Duration(value) * pattern.unit
				foundMatch = true
			}
		}
	}

	if !foundMatch {
		return 0, fmt.Errorf("invalid duration format: %s", durationStr)
	}

	return duration, nil
}
