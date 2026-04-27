package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds the runtime configuration for vaultpull.
type Config struct {
	// VaultAddr is the address of the Vault server.
	VaultAddr string

	// VaultToken is the token used to authenticate with Vault.
	VaultToken string

	// Namespace is the Vault namespace prefix to filter secrets.
	Namespace string

	// OutputFile is the path to the .env file to write secrets into.
	OutputFile string

	// MountPath is the KV secrets engine mount path (e.g. "secret").
	MountPath string
}

// Load builds a Config from environment variables and provided overrides.
// Explicit overrides (non-empty strings) take precedence over env vars.
func Load(overrides Config) (*Config, error) {
	cfg := &Config{
		VaultAddr:  firstNonEmpty(overrides.VaultAddr, os.Getenv("VAULT_ADDR")),
		VaultToken: firstNonEmpty(overrides.VaultToken, os.Getenv("VAULT_TOKEN")),
		Namespace:  firstNonEmpty(overrides.Namespace, os.Getenv("VAULT_NAMESPACE")),
		OutputFile: firstNonEmpty(overrides.OutputFile, ".env"),
		MountPath:  firstNonEmpty(overrides.MountPath, "secret"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Normalise namespace: strip leading/trailing slashes.
	cfg.Namespace = strings.Trim(cfg.Namespace, "/")

	return cfg, nil
}

// validate checks that required fields are present.
func (c *Config) validate() error {
	if c.VaultAddr == "" {
		return errors.New("vault address is required (set VAULT_ADDR or --vault-addr)")
	}
	if c.VaultToken == "" {
		return errors.New("vault token is required (set VAULT_TOKEN or --vault-token)")
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
