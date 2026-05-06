package sync

// Exported helpers for validator white-box testing.

// ValidationErrorExported re-exports ValidationError for use in external test packages.
type ValidationErrorExported = ValidationError

// NewValidatorExported exposes NewValidator for external tests.
var NewValidatorExported = NewValidator

// NewValidatingWriterExported exposes NewValidatingWriter for external tests.
var NewValidatingWriterExported = NewValidatingWriter

// NoEmptyKeyExported exposes the NoEmptyKey rule.
var NoEmptyKeyExported = NoEmptyKey

// NoEmptyValueExported exposes the NoEmptyValue rule.
var NoEmptyValueExported = NoEmptyValue

// MaxValueLengthExported exposes the MaxValueLength rule factory.
var MaxValueLengthExported = MaxValueLength
