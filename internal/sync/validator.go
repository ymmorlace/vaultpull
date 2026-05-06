package sync

import (
	"context"
	"fmt"
	"strings"
)

// ValidationError represents a secret key/value validation failure.
type ValidationError struct {
	Key    string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for key %q: %s", e.Key, e.Reason)
}

// ValidatorFunc is a function that validates a key-value pair.
// It returns an error if validation fails.
type ValidatorFunc func(key, value string) error

// Validator holds a set of validation rules applied before writing.
type Validator struct {
	rules []ValidatorFunc
}

// NewValidator creates a Validator with the provided rules.
func NewValidator(rules ...ValidatorFunc) *Validator {
	return &Validator{rules: rules}
}

// Validate runs all rules against key and value.
// Returns the first validation error encountered, or nil.
func (v *Validator) Validate(key, value string) error {
	for _, rule := range v.rules {
		if err := rule(key, value); err != nil {
			return err
		}
	}
	return nil
}

// NoEmptyKey rejects blank keys.
func NoEmptyKey(key, _ string) error {
	if strings.TrimSpace(key) == "" {
		return &ValidationError{Key: key, Reason: "key must not be empty"}
	}
	return nil
}

// NoEmptyValue rejects blank values.
func NoEmptyValue(key, value string) error {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{Key: key, Reason: "value must not be empty"}
	}
	return nil
}

// MaxValueLength rejects values exceeding maxLen bytes.
func MaxValueLength(maxLen int) ValidatorFunc {
	return func(key, value string) error {
		if len(value) > maxLen {
			return &ValidationError{Key: key, Reason: fmt.Sprintf("value exceeds maximum length of %d", maxLen)}
		}
		return nil
	}
}

// ValidatingWriter wraps an EnvWriter and validates each entry before writing.
type ValidatingWriter struct {
	inner     EnvWriter
	validator *Validator
}

// NewValidatingWriter creates a ValidatingWriter.
func NewValidatingWriter(inner EnvWriter, validator *Validator) *ValidatingWriter {
	if inner == nil {
		panic("validatingwriter: inner writer must not be nil")
	}
	if validator == nil {
		panic("validatingwriter: validator must not be nil")
	}
	return &ValidatingWriter{inner: inner, validator: validator}
}

// Write validates the entry and, if valid, delegates to the inner writer.
func (w *ValidatingWriter) Write(ctx context.Context, key, value string) error {
	if err := w.validator.Validate(key, value); err != nil {
		return err
	}
	return w.inner.Write(ctx, key, value)
}
