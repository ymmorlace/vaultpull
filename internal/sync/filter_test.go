package sync

import (
	"testing"
)

func TestNamespaceFilter_EmptyPrefix(t *testing.T) {
	f := NewNamespaceFilter("")
	paths := []string{"app/db", "infra/redis", "team/secret"}
	got := f.FilterPaths(paths)
	if len(got) != len(paths) {
		t.Errorf("expected %d paths, got %d", len(paths), len(got))
	}
}

func TestNamespaceFilter_Match(t *testing.T) {
	tests := []struct {
		prefix string
		path   string
		want   bool
	}{
		{"app", "app/db", true},
		{"app/", "app/db", true},
		{"app", "infra/db", false},
		{"app", "application/db", false},
		{"", "anything/goes", true},
	}

	for _, tt := range tests {
		t.Run(tt.prefix+":"+tt.path, func(t *testing.T) {
			f := NewNamespaceFilter(tt.prefix)
			if got := f.Match(tt.path); got != tt.want {
				t.Errorf("Match(%q) with prefix %q = %v, want %v", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestNamespaceFilter_Strip(t *testing.T) {
	tests := []struct {
		prefix string
		path   string
		want   string
	}{
		{"app", "app/db/password", "db/password"},
		{"app/", "app/db/password", "db/password"},
		{"app", "infra/db", "infra/db"},
		{"", "app/db", "app/db"},
	}

	for _, tt := range tests {
		t.Run(tt.prefix+":"+tt.path, func(t *testing.T) {
			f := NewNamespaceFilter(tt.prefix)
			if got := f.Strip(tt.path); got != tt.want {
				t.Errorf("Strip(%q) with prefix %q = %q, want %q", tt.path, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestNamespaceFilter_FilterPaths(t *testing.T) {
	f := NewNamespaceFilter("team")
	input := []string{"team/alpha/key", "team/beta/key", "infra/db", "app/secret"}
	got := f.FilterPaths(input)
	if len(got) != 2 {
		t.Fatalf("expected 2 filtered paths, got %d: %v", len(got), got)
	}
	if got[0] != "team/alpha/key" || got[1] != "team/beta/key" {
		t.Errorf("unexpected filtered paths: %v", got)
	}
}
