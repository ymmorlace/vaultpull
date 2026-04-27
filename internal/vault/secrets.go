package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SecretData holds the key-value pairs for a single secret.
type SecretData map[string]string

// ReadSecret fetches a KV v2 secret at the given path and returns its data fields.
func (c *Client) ReadSecret(path string) (SecretData, error) {
	path = strings.TrimPrefix(path, "/")
	url := fmt.Sprintf("%s/v1/secret/data/%s", c.addr, path)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	data := make(SecretData, len(result.Data.Data))
	for k, v := range result.Data.Data {
		data[k] = fmt.Sprintf("%v", v)
	}
	return data, nil
}
