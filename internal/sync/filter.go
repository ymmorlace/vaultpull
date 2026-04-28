package sync

import "strings"

// NamespaceFilter filters secret paths based on a namespace prefix.
type NamespaceFilter struct {
	prefix string
}

// NewNamespaceFilter creates a new NamespaceFilter with the given prefix.
// An empty prefix means all secrets pass through.
func NewNamespaceFilter(prefix string) *NamespaceFilter {
	return &NamespaceFilter{prefix: prefix}
}

// Match returns true if the given path matches the namespace prefix.
func (f *NamespaceFilter) Match(path string) bool {
	if f.prefix == "" {
		return true
	}
	normPrefix := strings.TrimSuffix(f.prefix, "/") + "/"
	return strings.HasPrefix(path, normPrefix)
}

// Strip removes the namespace prefix from the path, returning the relative path.
// If the path does not match the prefix, the original path is returned.
func (f *NamespaceFilter) Strip(path string) string {
	if f.prefix == "" {
		return path
	}
	normPrefix := strings.TrimSuffix(f.prefix, "/") + "/"
	if strings.HasPrefix(path, normPrefix) {
		return strings.TrimPrefix(path, normPrefix)
	}
	return path
}

// FilterPaths returns only the paths that match the namespace prefix.
func (f *NamespaceFilter) FilterPaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		if f.Match(p) {
			result = append(result, p)
		}
	}
	return result
}
