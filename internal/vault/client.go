package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods.
type Client struct {
	api    *vaultapi.Client
	prefix string
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(addr, token, prefix string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}
	api.SetToken(token)

	return &Client{api: api, prefix: prefix}, nil
}

// ListSecrets returns all secret key-value pairs under the given namespace path.
// It performs a recursive walk starting from the namespace.
func (c *Client) ListSecrets(ctx context.Context, namespace string) (map[string]string, error) {
	path := strings.Trim(namespace, "/")
	if c.prefix != "" {
		path = strings.Trim(c.prefix, "/") + "/" + path
	}

	secrets := make(map[string]string)
	if err := c.walk(ctx, path, secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

// walk recursively traverses the KV secret engine paths.
func (c *Client) walk(ctx context.Context, path string, acc map[string]string) error {
	// Try reading as a secret first.
	secret, err := c.api.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("reading path %q: %w", path, err)
	}

	if secret != nil && secret.Data != nil {
		if data, ok := secret.Data["data"]; ok {
			// KV v2
			if kvMap, ok := data.(map[string]interface{}); ok {
				for k, v := range kvMap {
					acc[k] = fmt.Sprintf("%v", v)
				}
				return nil
			}
		}
		// KV v1
		for k, v := range secret.Data {
			if k == "keys" {
				continue
			}
			acc[k] = fmt.Sprintf("%v", v)
		}
		return nil
	}

	// Try listing (directory node).
	list, err := c.api.Logical().ListWithContext(ctx, path)
	if err != nil {
		return fmt.Errorf("listing path %q: %w", path, err)
	}
	if list == nil {
		return nil
	}

	keys, ok := list.Data["keys"].([]interface{})
	if !ok {
		return nil
	}
	for _, key := range keys {
		child := path + "/" + strings.TrimSuffix(fmt.Sprintf("%v", key), "/")
		if err := c.walk(ctx, child, acc); err != nil {
			return err
		}
	}
	return nil
}
