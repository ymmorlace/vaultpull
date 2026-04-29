package sync_test

import (
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

func TestTrimSpaceTransformer(t *testing.T) {
	tr := sync.NewTrimSpaceTransformer()

	cases := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"\tvalue\n", "value"},
		{"clean", "clean"},
		{"", ""},
	}

	for _, tc := range cases {
		got, err := tr.Transform("KEY", tc.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tc.expected {
			t.Errorf("Transform(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestQuoteTransformer_NoQuoteNeeded(t *testing.T) {
	tr := sync.NewQuoteTransformer()
	got, err := tr.Transform("KEY", "simplevalue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "simplevalue" {
		t.Errorf("expected no quoting, got %q", got)
	}
}

func TestQuoteTransformer_QuotesSpaces(t *testing.T) {
	tr := sync.NewQuoteTransformer()
	got, err := tr.Transform("KEY", "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `"hello world"` {
		t.Errorf("expected quoted value, got %q", got)
	}
}

func TestQuoteTransformer_EscapesInnerQuotes(t *testing.T) {
	tr := sync.NewQuoteTransformer()
	got, err := tr.Transform("KEY", `say "hi" now`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `"say \"hi\" now"`
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestChainTransformer_AppliesInOrder(t *testing.T) {
	chain := sync.NewChainTransformer(
		sync.NewTrimSpaceTransformer(),
		sync.NewQuoteTransformer(),
	)

	got, err := chain.Transform("KEY", "  hello world  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != `"hello world"` {
		t.Errorf("got %q, want %q", got, `"hello world"`)
	}
}

func TestChainTransformer_EmptyChain(t *testing.T) {
	chain := sync.NewChainTransformer()
	got, err := chain.Transform("KEY", "value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "value" {
		t.Errorf("expected unchanged value, got %q", got)
	}
}
