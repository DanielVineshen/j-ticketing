// FILE: pkg/util/password_generator.go
package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Constants for password generation
const (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

// GenerateRandomPassword generates a random password with the specified length
func GenerateRandomPassword(length int) (string, error) {
	if length < 8 {
		// Enforce minimum security
		length = 8
	}

	// Ensure we have at least one of each character type
	password := make([]byte, length)

	// Add at least one lowercase letter
	randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return "", err
	}
	randomChar, err := rand.Int(rand.Reader, big.NewInt(int64(len(lowerCharSet))))
	if err != nil {
		return "", err
	}
	password[randomIndex.Int64()] = lowerCharSet[randomChar.Int64()]

	// Add at least one uppercase letter
	randomIndex, err = rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return "", err
	}
	randomChar, err = rand.Int(rand.Reader, big.NewInt(int64(len(upperCharSet))))
	if err != nil {
		return "", err
	}
	password[randomIndex.Int64()] = upperCharSet[randomChar.Int64()]

	// Add at least one number
	randomIndex, err = rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return "", err
	}
	randomChar, err = rand.Int(rand.Reader, big.NewInt(int64(len(numberSet))))
	if err != nil {
		return "", err
	}
	password[randomIndex.Int64()] = numberSet[randomChar.Int64()]

	// Add at least one special character
	randomIndex, err = rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return "", err
	}
	randomChar, err = rand.Int(rand.Reader, big.NewInt(int64(len(specialCharSet))))
	if err != nil {
		return "", err
	}
	password[randomIndex.Int64()] = specialCharSet[randomChar.Int64()]

	// Fill the rest with random characters
	for i := 0; i < length; i++ {
		if password[i] == 0 {
			randomChar, err = rand.Int(rand.Reader, big.NewInt(int64(len(allCharSet))))
			if err != nil {
				return "", err
			}
			password[i] = allCharSet[randomChar.Int64()]
		}
	}

	return string(password), nil
}

// GenerateCustomerID generates a unique customer ID with a prefix
func GenerateCustomerID(prefix string) (string, error) {
	// Generate a random 8-character alphanumeric string
	const idCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const idLength = 8

	id := make([]byte, idLength)
	for i := 0; i < idLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(idCharSet))))
		if err != nil {
			return "", err
		}
		id[i] = idCharSet[randomIndex.Int64()]
	}

	// Combine prefix and random string
	return fmt.Sprintf("%s-%s", prefix, string(id)), nil
}
