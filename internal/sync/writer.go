package sync

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// EnvWriter writes a map of key/value pairs to a destination.
type EnvWriter interface {
	Write(path string, secrets map[string]string) error
}

// FileEnvWriter writes secrets to a .env file on disk.
type FileEnvWriter struct{}

// Write serialises secrets as KEY=VALUE lines and atomically writes them to
// the file at path, creating or truncating it as needed.
func (w FileEnvWriter) Write(path string, secrets map[string]string) error {
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		v := secrets[k]
		// Quote values that contain whitespace or special characters.
		if strings.ContainsAny(v, " \t\n\r#") {
			v = fmt.Sprintf("%q", v)
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		sb.WriteByte('\n')
	}

	return os.WriteFile(path, []byte(sb.String()), 0o600)
}
