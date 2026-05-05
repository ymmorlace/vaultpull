package sync_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/your-org/vaultpull/internal/sync"
)

// recordingWriter captures every Write call for assertion.
type recordingWriter struct {
	entries []string
	errAfter int // return error after this many writes (-1 = never)
	calls    int
}

func (r *recordingWriter) Write(_ context.Context, key, value string) error {
	r.calls++
	if r.errAfter >= 0 && r.calls > r.errAfter {
		return errors.New("recording writer: forced error")
	}
	r.entries = append(r.entries, fmt.Sprintf("%s=%s", key, value))
	return nil
}

func TestPipeline_RunsAllStages(t *testing.T) {
	a := &recordingWriter{errAfter: -1}
	b := &recordingWriter{errAfter: -1}

	p := sync.NewPipeline(
		sync.Stage{Name: "a", Writer: a},
		sync.Stage{Name: "b", Writer: b},
	)

	if err := p.Run(context.Background(), "KEY", "val"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.entries) != 1 || a.entries[0] != "KEY=val" {
		t.Errorf("stage a: got %v", a.entries)
	}
	if len(b.entries) != 1 || b.entries[0] != "KEY=val" {
		t.Errorf("stage b: got %v", b.entries)
	}
}

func TestPipeline_HaltsOnStageError(t *testing.T) {
	a := &recordingWriter{errAfter: 0} // errors on first write
	b := &recordingWriter{errAfter: -1}

	p := sync.NewPipeline(
		sync.Stage{Name: "fail", Writer: a},
		sync.Stage{Name: "skip", Writer: b},
	)

	err := p.Run(context.Background(), "K", "v")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errors.New("recording writer: forced error")) {
		// just check wrapping message
		if fmt.Sprintf("%v", err) == "" {
			t.Errorf("error message is empty")
		}
	}
	if b.calls != 0 {
		t.Errorf("second stage should not have been called, got %d calls", b.calls)
	}
}

func TestPipeline_StageNames(t *testing.T) {
	p := sync.NewPipeline(
		sync.Stage{Name: "transform", Writer: &recordingWriter{errAfter: -1}},
		sync.Stage{Name: "write", Writer: &recordingWriter{errAfter: -1}},
	)
	names := p.StageNames()
	if len(names) != 2 || names[0] != "transform" || names[1] != "write" {
		t.Errorf("unexpected stage names: %v", names)
	}
}

func TestPipeline_PanicOnNoStages(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty stages")
		}
	}()
	sync.NewPipeline()
}
