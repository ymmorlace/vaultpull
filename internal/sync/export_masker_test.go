package sync

// Exported helpers for white-box masker tests.

// NewMaskerExported exposes NewMasker to external test packages.
func NewMaskerExported(mode MaskMode, reveal int, maskChar rune) *Masker {
	return NewMasker(mode, reveal, maskChar)
}

// MaskModeFullExported exposes MaskFull to external test packages.
const MaskModeFullExported = MaskFull

// MaskModePartialExported exposes MaskPartial to external test packages.
const MaskModePartialExported = MaskPartial
