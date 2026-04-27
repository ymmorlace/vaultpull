package config

import (
	"os"
	"testing"
)

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "root")
	t.Setenv("VAULT_NAMESPACE", "/team/backend/")

	cfg, err := Load(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.VaultAddr != "http://127.0.0.1:8200" {
		t.Errorf("VaultAddr = %q, want %q", cfg.VaultAddr, "http://127.0.0.1:8200")
	}
	if cfg.VaultToken != "root" {
		t.Errorf("VaultToken = %q, want %q", cfg.VaultToken, "root")
	}
	// Namespace should have slashes stripped.
	if cfg.Namespace != "team/backend" {
		t.Errorf("Namespace = %q, want %q", cfg.Namespace, "team/backend")
	}
}

func TestLoad_OverridesTakePrecedence(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://env-addr:8200")
	t.Setenv("VAULT_TOKEN", "env-token")

	cfg, err := Load(Config{
		VaultAddr:  "http://override-addr:8200",
		VaultToken: "override-token",
		OutputFile: "prod.env",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.VaultAddr != "http://override-addr:8200" {
		t.Errorf("VaultAddr = %q, want override value", cfg.VaultAddr)
	}
	if cfg.OutputFile != "prod.env" {
		t.Errorf("OutputFile = %q, want %q", cfg.OutputFile, "prod.env")
	}
}

func TestLoad_DefaultOutputFile(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	t.Setenv("VAULT_TOKEN", "root")

	cfg, err := Load(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputFile != ".env" {
		t.Errorf("OutputFile = %q, want \".env\"", cfg.OutputFile)
	}
	if cfg.MountPath != "secret" {
		t.Errorf("MountPath = %q, want \"secret\"", cfg.MountPath)
	}
}

func TestLoad_MissingAddr(t *testing.T) {
	os.Unsetenv("VAULT_ADDR")
	t.Setenv("VAULT_TOKEN", "root")

	_, err := Load(Config{})
	if err == nil {
		t.Fatal("expected error for missing VAULT_ADDR, got nil")
	}
}

func TestLoad_MissingToken(t *testing.T) {
	t.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	os.Unsetenv("VAULT_TOKEN")

	_, err := Load(Config{})
	if err == nil {
		t.Fatal("expected error for missing VAULT_TOKEN, got nil")
	}
}
