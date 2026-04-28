package sync

import (
	"fmt"
	"log"

	"github.com/user/vaultpull/internal/vault"
)

// Runner orchestrates the full sync process: list, filter, read, and write secrets.
type Runner struct {
	client  *vault.Client
	filter  *NamespaceFilter
	syncer  *Syncer
	writer  EnvWriter
	mount   string
}

// NewRunner constructs a Runner with the provided dependencies.
func NewRunner(client *vault.Client, filter *NamespaceFilter, syncer *Syncer, writer EnvWriter, mount string) *Runner {
	return &Runner{
		client: client,
		filter: filter,
		syncer: syncer,
		writer: writer,
		mount:  mount,
	}
}

// Run executes the full sync pipeline.
func (r *Runner) Run() error {
	paths, err := r.client.ListSecrets(r.mount, "")
	if err != nil {
		return fmt.Errorf("listing secrets: %w", err)
	}

	filtered := r.filter.FilterPaths(paths)
	if len(filtered) == 0 {
		log.Println("no secrets matched the namespace filter")
		return nil
	}

	secrets := make(map[string]string, len(filtered))
	for _, path := range filtered {
		stripped := r.filter.Strip(path)
		data, err := r.client.ReadSecret(r.mount, path)
		if err != nil {
			log.Printf("warn: skipping %s: %v", path, err)
			continue
		}
		for k, v := range data {
			envKey := r.syncer.ToEnvKey(stripped + "/" + k)
			secrets[envKey] = fmt.Sprintf("%v", v)
		}
	}

	if err := r.writer.Write(secrets); err != nil {
		return fmt.Errorf("writing env file: %w", err)
	}

	log.Printf("synced %d secret keys", len(secrets))
	return nil
}
