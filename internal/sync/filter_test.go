package sync

import (
	"testing"
)

func TestNamespaceFilter_EmptyPrefix(t *testing.T) {
	f := NewNamespaceFilter("")
	paths := []string{"app/db", "infra/redis", "team/secret"}
	got := f.FilterPaths(paths)
	if len(got) != len(paths) {
		t.Fatalf("expected %d paths, got %d", len(paths), len(got))
	}
	for i, p := range paths {
		if got[i] != p {
			t.Errorf("expected %q, got %q", p, got[i])
		}
	}
}

func TestNamespaceFilter_Match(t *testing.T) {
	f := NewNamespaceFilter("app")
	tests := []struct {
		path  string
		want  bool
	}{
		{"app/db", true},
		{"app/cache", true},
		{"infra/redis", false},
		{"application/x", true}, // HasPrefix match
		{"", false},
	}
	for _, tt := range tests {
		got := f.Match(tt.path)
		if got != tt.want {
			t.Errorf("Match(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestNamespaceFilter_Strip(t *testing.T) {
	f := NewNamespaceFilter("app/")
	tests := []struct {
		path string
		want string
	}{
		{"app/db", "db"},
		{"app/cache/redis", "cache/redis"},
		{"infra/redis", "infra/redis"},
		{"app/", ""},
	}
	for _, tt := range tests {
		got := f.Strip(tt.path)
		if got != tt.want {
			t.Errorf("Strip(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestNamespaceFilter_FilterPaths(t *testing.T) {
	f := NewNamespaceFilter("app/")
	input := []string{"app/db", "app/cache", "infra/redis", "app/secrets"}
	want := []string{"db", "cache", "secrets"}
	got := f.FilterPaths(input)
	if len(got) != len(want) {
		t.Fatalf("expected %d results, got %d: %v", len(want), len(got), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: expected %q, got %q", i, w, got[i])
		}
	}
}
