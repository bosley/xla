package xrt

import (
	"errors"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

// NewUser creates a new User with the given username, password, email, and isAgent status
func NewUser(username, password, email string, isAgent bool) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	if !isValidEmail(email) {
		return nil, errors.New("invalid email format")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create a new User
	user := &User{
		Name:         username,
		PasswordHash: hashedPassword,
		Email:        email,
		IsAgent:      isAgent,
	}

	return user, nil
}

// isValidEmail checks if the given email is valid using the net/mail package
func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
