package sync_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultpull/internal/sync"
)

func TestPipelineBuilder_BuildsWithAdd(t *testing.T) {
	w := &recordingWriter{errAfter: -1}
	p := sync.NewPipelineBuilder().
		Add("primary", w).
		Build()

	if err := p.Run(context.Background(), "FOO", "bar"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.entries) != 1 || w.entries[0] != "FOO=bar" {
		t.Errorf("got entries: %v", w.entries)
	}
}

func TestPipelineBuilder_WithAuditLogsEntry(t *testing.T) {
	var buf bytes.Buffer
	w := &recordingWriter{errAfter: -1}

	p := sync.NewPipelineBuilder().
		WithAudit("audit", &buf, w).
		Build()

	if err := p.Run(context.Background(), "SECRET", "value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "SECRET") {
		t.Errorf("audit log missing key, got: %s", buf.String())
	}
}

func TestPipelineBuilder_WithTimeout_Completes(t *testing.T) {
	w := &recordingWriter{errAfter: -1}

	p := sync.NewPipelineBuilder().
		WithTimeout("timeout", 2*time.Second, w).
		Build()

	if err := p.Run(context.Background(), "K", "v"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipelineBuilder_StageNamesPreserved(t *testing.T) {
	w := &recordingWriter{errAfter: -1}

	p := sync.NewPipelineBuilder().
		Add("first", w).
		Add("second", w).
		Build()

	names := p.StageNames()
	if len(names) != 2 || names[0] != "first" || names[1] != "second" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestPipelineBuilder_PanicOnEmptyBuild(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty builder")
		}
	}()
	sync.NewPipelineBuilder().Build()
}
