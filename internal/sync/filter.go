package sync

import "strings"

// NamespaceFilter filters and transforms secret paths based on a namespace prefix.
type NamespaceFilter struct {
	prefix string
}

// NewNamespaceFilter creates a new NamespaceFilter with the given prefix.
// An empty prefix means all paths are accepted without stripping.
func NewNamespaceFilter(prefix string) *NamespaceFilter {
	return &NamespaceFilter{prefix: prefix}
}

// Match reports whether the given path matches the namespace prefix.
// If the prefix is empty, all paths match.
func (f *NamespaceFilter) Match(path string) bool {
	if f.prefix == "" {
		return true
	}
	return strings.HasPrefix(path, f.prefix)
}

// Strip removes the namespace prefix from the path.
// If the prefix is empty or the path does not start with it, the original path is returned.
func (f *NamespaceFilter) Strip(path string) string {
	if f.prefix == "" {
		return path
	}
	trimmed := strings.TrimPrefix(path, f.prefix)
	return strings.TrimPrefix(trimmed, "/")
}

// FilterPaths returns only the paths that match the prefix, with the prefix stripped.
func (f *NamespaceFilter) FilterPaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		if f.Match(p) {
			result = append(result, f.Strip(p))
		}
	}
	return result
}
