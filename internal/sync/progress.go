package sync

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ProgressReporter tracks and displays sync progress to a writer.
type ProgressReporter struct {
	mu        sync.Mutex
	out       io.Writer
	total     int
	done      int
	skipped   int
	errored   int
	start     time.Time
	clock     func() time.Time
}

// NewProgressReporter creates a ProgressReporter writing to out.
// If out is nil, os.Stderr is used.
func NewProgressReporter(out io.Writer, total int) *ProgressReporter {
	if out == nil {
		out = os.Stderr
	}
	return &ProgressReporter{
		out:   out,
		total: total,
		start: time.Now(),
		clock: time.Now,
	}
}

// RecordWritten increments the written counter and prints progress.
func (p *ProgressReporter) RecordWritten(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done++
	p.print(fmt.Sprintf("wrote   %s", key))
}

// RecordSkipped increments the skipped counter.
func (p *ProgressReporter) RecordSkipped(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.skipped++
	p.print(fmt.Sprintf("skipped %s", key))
}

// RecordError increments the errored counter.
func (p *ProgressReporter) RecordError(key string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.errored++
	p.print(fmt.Sprintf("error   %s: %v", key, err))
}

// Summary prints a final summary line.
func (p *ProgressReporter) Summary() {
	p.mu.Lock()
	defer p.mu.Unlock()
	elapsed := p.clock().Sub(p.start).Round(time.Millisecond)
	fmt.Fprintf(p.out, "sync complete: %d written, %d skipped, %d errors in %s\n",
		p.done, p.skipped, p.errored, elapsed)
}

// Counts returns (written, skipped, errored).
func (p *ProgressReporter) Counts() (int, int, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.done, p.skipped, p.errored
}

func (p *ProgressReporter) print(msg string) {
	if p.total > 0 {
		current := p.done + p.skipped + p.errored
		fmt.Fprintf(p.out, "[%d/%d] %s\n", current, p.total, msg)
	} else {
		fmt.Fprintf(p.out, "%s\n", msg)
	}
}
