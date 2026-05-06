package sync_test

import (
	"context"
)

// captureWriter is a test helper that records Write calls.
type captureWriter struct {
	entries []struct{ key, value string }
	err     error
}

func (c *captureWriter) Write(_ context.Context, key, value string) error {
	if c.err != nil {
		return c.err
	}
	c.entries = append(c.entries, struct{ key, value string }{key, value})
	return nil
}
