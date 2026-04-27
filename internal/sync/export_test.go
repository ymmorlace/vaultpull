package sync

// ExportToEnvKey exposes the unexported toEnvKey function for white-box
// testing from the sync_test package.
func ExportToEnvKey(keyPath, field string) string {
	return toEnvKey(keyPath, field)
}
