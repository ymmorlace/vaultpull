package main

import (
	"fmt"
	"os"

	"github.com/vaultpull/internal/config"
	"github.com/vaultpull/internal/sync"
	"github.com/vaultpull/internal/vault"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("vaultpull %s\n", version)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	client, err := vault.NewClient(cfg.VaultAddr, cfg.VaultToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating vault client: %v\n", err)
		os.Exit(1)
	}

	writer := sync.NewFileEnvWriter(cfg.OutputFile)
	filter := sync.NewNamespaceFilter(cfg.Namespace)
	runner := sync.NewRunner(client, writer, filter)

	if err := runner.Run(cfg.SecretPath); err != nil {
		fmt.Fprintf(os.Stderr, "error syncing secrets: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("secrets synced to %s\n", cfg.OutputFile)
}
