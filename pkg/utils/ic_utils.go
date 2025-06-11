package utils

import (
	"strconv"
	"strings"
)

// ExtractAgeFromMalaysianIC extracts age from Malaysian IC number
func ExtractAgeFromMalaysianIC(ic string, currentYear int) int {
	isMalaysianIC := IsMalaysianIC(ic)

	if isMalaysianIC == false {
		return -1 // Invalid IC
	}

	// Get first two digits
	yearStr := ic[:2]

	// Convert to integer
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return -1 // Invalid year
	}

	// Determine full birth year
	// Malaysian IC format:
	// - 00-99 for years 1900-1999 (old format)
	// - For new format, need to check century
	// Generally, if year > current year's last 2 digits, it's from previous century

	var birthYear int
	currentYearLastTwo := currentYear % 100

	if year > currentYearLastTwo {
		// Previous century (1900s)
		birthYear = 1900 + year
	} else {
		// Current century (2000s)
		birthYear = 2000 + year
	}

	// Calculate age
	age := currentYear - birthYear

	// Validate reasonable age range (0-150)
	if age < 0 || age > 150 {
		return -1 // Invalid age
	}

	return age
}

// CategorizeAge categorizes age into predefined groups
func CategorizeAge(age int) string {
	if age < 0 {
		return "Unknown"
	}

	switch {
	case age <= 12:
		return "0-12"
	case age <= 17:
		return "13-17"
	case age <= 35:
		return "18-35"
	case age <= 50:
		return "36-50"
	default:
		return "51+"
	}
}

// DetermineNationality determines if the visitor is local (Malaysian) or international
func DetermineNationality(identificationNo string) string {
	if IsMalaysianIC(identificationNo) {
		return "Local"
	}
	return "International"
}

// IsMalaysianIC checks if the identification number follows Malaysian IC format
func IsMalaysianIC(ic string) bool {
	// Remove any spaces or dashes
	ic = strings.ReplaceAll(ic, " ", "")
	ic = strings.ReplaceAll(ic, "-", "")

	// Malaysian IC can be 12 digits (original) or 14 digits (with replacement suffix)
	if len(ic) != 12 && len(ic) != 14 {
		return false
	}

	// Check if all characters are digits
	for _, char := range ic {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Extract date parts (first 6 digits: YYMMDD)
	month := ic[2:4]
	day := ic[4:6]

	// Strict date validation
	monthInt, err1 := strconv.Atoi(month)
	dayInt, err2 := strconv.Atoi(day)

	if err1 != nil || err2 != nil {
		return false
	}

	// Check valid month (01-12)
	if monthInt < 1 || monthInt > 12 {
		return false
	}

	// Check valid day (01-31)
	if dayInt < 1 || dayInt > 31 {
		return false
	}

	// Check place of birth code (positions 7-8)
	placeCode := ic[6:8]
	placeInt, err := strconv.Atoi(placeCode)
	if err != nil {
		return false
	}

	// Official Malaysian state/place codes (https://www.jpn.gov.my/my/kod-negeri)
	validPlaceCodes := []int{
		// Johor
		1, 21, 22, 23, 24,
		// Kedah
		2, 25, 26, 27,
		// Kelantan
		3, 28, 29,
		// Melaka
		4, 30,
		// Negeri Sembilan
		5, 31, 59,
		// Pahang
		6, 32, 33,
		// Pulau Pinang
		7, 34, 35,
		// Perak
		8, 36, 37, 38, 39,
		// Perlis
		9, 40,
		// Selangor
		10, 41, 42, 43, 44,
		// Terengganu
		11, 45, 46,
		// Sabah
		12, 47, 48, 49,
		// Sarawak
		13, 50, 51, 52, 53,
		// Wilayah Persekutuan (Kuala Lumpur)
		14, 54, 55, 56, 57,
		// Wilayah Persekutuan (Labuan)
		15, 58,
		// Wilayah Persekutuan (Putrajaya)
		16,
		// Negeri Tidak Diketahui
		82,
	}

	// Check if place code is valid
	isValidPlace := false
	for _, validCode := range validPlaceCodes {
		if placeInt == validCode {
			isValidPlace = true
			break
		}
	}

	if !isValidPlace {
		return false
	}

	// If 14 digits, validate the replacement suffix (last 2 digits should be 01-99)
	if len(ic) == 14 {
		replacementSuffix := ic[12:14]
		suffixInt, err := strconv.Atoi(replacementSuffix)
		if err != nil || suffixInt < 1 || suffixInt > 99 {
			return false
		}
	}

	return true
}
