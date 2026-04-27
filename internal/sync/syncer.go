package sync

import (
	"fmt"
	"strings"

	"github.com/vaultpull/internal/vault"
)

// Syncer pulls secrets from Vault and writes them to an env file.
type Syncer struct {
	client    *vault.Client
	namespace string
	output    string
	writer    EnvWriter
}

// NewSyncer creates a Syncer with the given Vault client, namespace filter,
// output file path, and writer.
func NewSyncer(client *vault.Client, namespace, output string, writer EnvWriter) *Syncer {
	return &Syncer{
		client:    client,
		namespace: namespace,
		output:    output,
		writer:    writer,
	}
}

// Run fetches secrets matching the namespace prefix and writes them to the
// configured output file.
func (s *Syncer) Run(mountPath string) error {
	keys, err := s.client.ListSecrets(mountPath, s.namespace)
	if err != nil {
		return fmt.Errorf("listing secrets: %w", err)
	}

	secrets := make(map[string]string, len(keys))
	for _, key := range keys {
		path := strings.TrimSuffix(mountPath, "/") + "/" + strings.TrimPrefix(key, "/")
		data, err := s.client.GetSecret(path)
		if err != nil {
			return fmt.Errorf("fetching secret %q: %w", path, err)
		}
		for k, v := range data {
			envKey := toEnvKey(key, k)
			secrets[envKey] = fmt.Sprintf("%v", v)
		}
	}

	return s.writer.Write(s.output, secrets)
}

// toEnvKey converts a vault key path and field name into an upper-snake-case
// environment variable name.
func toEnvKey(keyPath, field string) string {
	base := strings.ReplaceAll(keyPath, "/", "_")
	base = strings.ReplaceAll(base, "-", "_")
	return strings.ToUpper(base + "_" + field)
}
