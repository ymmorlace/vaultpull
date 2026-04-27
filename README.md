# vaultpull

> CLI tool to sync HashiCorp Vault secrets into local `.env` files with namespace filtering

---

## Installation

```bash
go install github.com/yourusername/vaultpull@latest
```

Or download a prebuilt binary from the [releases page](https://github.com/yourusername/vaultpull/releases).

---

## Usage

Authenticate with Vault and pull secrets into a `.env` file:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.xxxxxxxx"

vaultpull --namespace secret/myapp --output .env
```

**Flags:**

| Flag | Description | Default |
|-------------|-------------------------------|------------|
| `--namespace` | Vault secret path/namespace | `secret/` |
| `--output` | Output `.env` file path | `.env` |
| `--overwrite` | Overwrite existing file | `false` |

**Example output (`.env`):**

```env
DATABASE_URL=postgres://user:pass@localhost/db
API_KEY=abc123
DEBUG=false
```

---

## Requirements

- Go 1.21+
- A running HashiCorp Vault instance
- A valid `VAULT_TOKEN` or other supported auth method

---

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

---

## License

[MIT](LICENSE)