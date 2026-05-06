package sync

import (
	"strings"
	"unicode/utf8"
)

// MaskMode controls how a secret value is masked.
type MaskMode int

const (
	// MaskFull replaces the entire value with asterisks.
	MaskFull MaskMode = iota
	// MaskPartial reveals the last N characters and masks the rest.
	MaskPartial
)

// Masker transforms secret values into masked representations for safe display.
type Masker struct {
	mode    MaskMode
	reveal  int
	maskChar rune
}

// NewMasker creates a Masker with the given mode.
// For MaskPartial, reveal controls how many trailing characters remain visible.
// maskChar is the character used for masking (e.g. '*').
func NewMasker(mode MaskMode, reveal int, maskChar rune) *Masker {
	if maskChar == 0 {
		maskChar = '*'
	}
	if reveal < 0 {
		reveal = 0
	}
	return &Masker{mode: mode, reveal: reveal, maskChar: maskChar}
}

// Mask returns a masked version of value.
func (m *Masker) Mask(value string) string {
	if value == "" {
		return ""
	}
	switch m.mode {
	case MaskPartial:
		return m.maskPartial(value)
	default:
		return m.maskFull(value)
	}
}

func (m *Masker) maskFull(value string) string {
	count := utf8.RuneCountInString(value)
	return strings.Repeat(string(m.maskChar), count)
}

func (m *Masker) maskPartial(value string) string {
	runes := []rune(value)
	total := len(runes)
	if m.reveal >= total {
		return strings.Repeat(string(m.maskChar), total)
	}
	hidden := total - m.reveal
	var sb strings.Builder
	sb.WriteString(strings.Repeat(string(m.maskChar), hidden))
	sb.WriteString(string(runes[hidden:]))
	return sb.String()
}

// DefaultMasker is a Masker that fully masks values using '*'.
var DefaultMasker = NewMasker(MaskFull, 0, '*')
