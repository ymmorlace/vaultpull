package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestMain_Version(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "version")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected no error running version command, got: %v", err)
	}

	output := strings.TrimSpace(string(out))
	if !strings.HasPrefix(output, "vaultpull ") {
		t.Errorf("expected output to start with 'vaultpull ', got: %q", output)
	}

	if !strings.Contains(output, version) {
		t.Errorf("expected output to contain version %q, got: %q", version, output)
	}
}

func TestMain_MissingConfig(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = []string{"HOME=/tmp"} // strip all vault env vars

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected error when vault config is missing, got none")
	}

	output := string(out)
	if !strings.Contains(output, "error") {
		t.Errorf("expected error message in output, got: %q", output)
	}
}
