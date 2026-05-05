package sync

import (
	"context"
	"fmt"
)

// Stage represents a single step in a sync pipeline.
type Stage struct {
	Name   string
	Writer EnvWriter
}

// Pipeline executes a sequence of stages, passing each secret through all
// writers in order. If any stage returns an error the pipeline halts.
type Pipeline struct {
	stages []Stage
}

// NewPipeline constructs a Pipeline from the provided stages.
// It panics if no stages are supplied.
func NewPipeline(stages ...Stage) *Pipeline {
	if len(stages) == 0 {
		panic("pipeline: at least one stage is required")
	}
	return &Pipeline{stages: stages}
}

// Run writes a single key/value pair through every stage in order.
func (p *Pipeline) Run(ctx context.Context, key, value string) error {
	for _, s := range p.stages {
		if err := s.Writer.Write(ctx, key, value); err != nil {
			return fmt.Errorf("pipeline stage %q: %w", s.Name, err)
		}
	}
	return nil
}

// StageNames returns the ordered list of stage names for introspection.
func (p *Pipeline) StageNames() []string {
	names := make([]string, len(p.stages))
	for i, s := range p.stages {
		names[i] = s.Name
	}
	return names
}
