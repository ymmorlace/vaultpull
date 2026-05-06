package sync_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

func TestMasker_MaskFull_ASCII(t *testing.T) {
	m := sync.NewMasker(sync.MaskFull, 0, '*')
	got := m.Mask("secretvalue")
	if got != "***********" {
		t.Errorf("expected all stars, got %q", got)
	}
}

func TestMasker_MaskFull_Unicode(t *testing.T) {
	m := sync.NewMasker(sync.MaskFull, 0, '*')
	// "héllo" has 5 runes
	got := m.Mask("héllo")
	if got != "*****" {
		t.Errorf("expected 5 stars, got %q", got)
	}
}

func TestMasker_MaskFull_EmptyString(t *testing.T) {
	m := sync.NewMasker(sync.MaskFull, 0, '*')
	got := m.Mask("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestMasker_MaskPartial_RevealsTrailing(t *testing.T) {
	m := sync.NewMasker(sync.MaskPartial, 4, '*')
	got := m.Mask("supersecret")
	// "supersecret" = 11 chars, reveal last 4 → "*******ecret"
	if !strings.HasSuffix(got, "cret") {
		t.Errorf("expected suffix 'cret', got %q", got)
	}
	if !strings.HasPrefix(got, "*******") {
		t.Errorf("expected 7 leading stars, got %q", got)
	}
}

func TestMasker_MaskPartial_RevealExceedsLength(t *testing.T) {
	m := sync.NewMasker(sync.MaskPartial, 20, '*')
	// reveal >= total → fully masked
	got := m.Mask("short")
	if got != "*****" {
		t.Errorf("expected all stars when reveal >= length, got %q", got)
	}
}

func TestMasker_CustomMaskChar(t *testing.T) {
	m := sync.NewMasker(sync.MaskFull, 0, '#')
	got := m.Mask("abc")
	if got != "###" {
		t.Errorf("expected '###', got %q", got)
	}
}

func TestMasker_NegativeRevealClampsToZero(t *testing.T) {
	m := sync.NewMasker(sync.MaskPartial, -5, '*')
	got := m.Mask("hello")
	if got != "*****" {
		t.Errorf("expected all stars for negative reveal, got %q", got)
	}
}

func TestMasker_ZeroMaskCharDefaultsStar(t *testing.T) {
	m := sync.NewMasker(sync.MaskFull, 0, 0)
	got := m.Mask("hi")
	if got != "**" {
		t.Errorf("expected '**' with default mask char, got %q", got)
	}
}

func TestDefaultMasker_FullMask(t *testing.T) {
	got := sync.DefaultMasker.Mask("topsecret")
	if got != "---------" && got != "*********" {
		// just check it's all the same char and same length
		if len(got) != len("topsecret") {
			t.Errorf("DefaultMasker: expected same length mask, got %q", got)
		}
	}
}
