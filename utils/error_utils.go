package utils

import (
	"errors"
	"fmt"
)

// WrapError wraps an error with a custom message
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// NewError creates a new error with a specific message
func NewError(message string) error {
	return errors.New(message)
}

// LogAndWrapError logs an error and wraps it with a custom message
func LogAndWrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	logMessage := fmt.Sprintf("%s: %v", message, err)
	LogError(logMessage)
	return WrapError(err, message)
}
