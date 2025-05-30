// File: j-ticketing/pkg/utils/date.go
package utils

import (
	"fmt"
	"time"

	// Import the timezone database to embed it in the binary
	_ "time/tzdata"
)

const (
	// Malaysia timezone
	MalaysiaTimezone = "Asia/Kuala_Lumpur"

	// Malaysia UTC offset in hours (UTC+8)
	MalaysiaOffsetHours = 8

	// Common date formats
	FullDateTimeFormat = "2006-01-02 15:04:05" // yyyy-MM-dd HH:mm:ss
	DateOnlyFormat     = "2006-01-02"          // yyyy-MM-dd
	MonthYearFormat    = "2006-01"             // yyyy-MM
)

// getMalaysiaLocation returns the Malaysia timezone location with a fallback mechanism
func getMalaysiaLocation() (*time.Location, error) {
	// Try to load from the embedded timezone database first
	location, err := time.LoadLocation(MalaysiaTimezone)
	if err == nil {
		return location, nil
	}

	// If that fails, use a fixed UTC+8 offset as fallback
	// Create a fixed location based on name and offset
	return time.FixedZone("Malaysia", MalaysiaOffsetHours*60*60), nil
}

// ToMalaysiaTime converts UTC time to Malaysia timezone
func ToMalaysiaTime(utcTime time.Time) (time.Time, error) {
	location, err := getMalaysiaLocation()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to determine Malaysia timezone: %w", err)
	}
	return utcTime.In(location), nil
}

// FormatToMalaysiaTime converts UTC time to Malaysia timezone with specified format
func FormatToMalaysiaTime(utcTime time.Time, format string) (string, error) {
	malaysiaTime, err := ToMalaysiaTime(utcTime)
	if err != nil {
		return "", err
	}
	return malaysiaTime.Format(format), nil
}

// ParseUTC parses a date string with the given format as UTC
func ParseUTC(dateStr string, format string) (time.Time, error) {
	return time.Parse(format, dateStr)
}

// ParseAndConvertToMalaysiaTime parses a date string in UTC and converts to Malaysia time
func ParseAndConvertToMalaysiaTime(dateStr string, format string) (time.Time, error) {
	utcTime, err := ParseUTC(dateStr, format)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date: %w", err)
	}

	return ToMalaysiaTime(utcTime)
}

// ParseAndFormatToMalaysiaTime parses a date string in UTC and formats it to Malaysia time
func ParseAndFormatToMalaysiaTime(dateStr string, inputFormat string, outputFormat string) (string, error) {
	malaysiaTime, err := ParseAndConvertToMalaysiaTime(dateStr, inputFormat)
	if err != nil {
		return "", err
	}

	return malaysiaTime.Format(outputFormat), nil
}

// GetCurrentMalaysiaTime gets the current time in Malaysia timezone
func GetCurrentMalaysiaTime() (time.Time, error) {
	return ToMalaysiaTime(time.Now().UTC())
}

// FormatCurrentMalaysiaTime gets the current time in Malaysia timezone with specified format
func FormatCurrentMalaysiaTime(format string) (string, error) {
	malaysiaTime, err := GetCurrentMalaysiaTime()
	if err != nil {
		return "", err
	}

	return malaysiaTime.Format(format), nil
}
