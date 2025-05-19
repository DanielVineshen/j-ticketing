// FILE: j-ticketing/internal/pkg/utils/parser.go
package utils

import "strconv"

// Helper function to safely parse a string to float
func ParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0 // Return default value if parsing fails
	}
	return f
}
