package sync

import (
	"io"
	"time"
)

// PipelineBuilder provides a fluent API for assembling a Pipeline with
// commonly used decorators.
type PipelineBuilder struct {
	stages []Stage
}

// NewPipelineBuilder returns an empty builder.
func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{}
}

// WithTransform appends a transforming writer stage.
func (b *PipelineBuilder) WithTransform(name string, t Transformer, inner EnvWriter) *PipelineBuilder {
	b.stages = append(b.stages, Stage{
		Name:   name,
		Writer: NewTransformingWriter(inner, t),
	})
	return b
}

// WithAudit appends an auditing writer stage.
func (b *PipelineBuilder) WithAudit(name string, w io.Writer, inner EnvWriter) *PipelineBuilder {
	log := NewAuditLog(w)
	b.stages = append(b.stages, Stage{
		Name:   name,
		Writer: NewAuditingWriter(inner, log),
	})
	return b
}

// WithTimeout appends a timeout-enforcing writer stage.
func (b *PipelineBuilder) WithTimeout(name string, d time.Duration, inner EnvWriter) *PipelineBuilder {
	b.stages = append(b.stages, Stage{
		Name:   name,
		Writer: NewTimeoutWriter(inner, d),
	})
	return b
}

// WithMetrics appends a metrics-recording writer stage.
func (b *PipelineBuilder) WithMetrics(name string, m *SyncMetrics, inner EnvWriter) *PipelineBuilder {
	b.stages = append(b.stages, Stage{
		Name:   name,
		Writer: NewMetricsWriter(inner, m),
	})
	return b
}

// Add appends an arbitrary stage.
func (b *PipelineBuilder) Add(name string, w EnvWriter) *PipelineBuilder {
	b.stages = append(b.stages, Stage{Name: name, Writer: w})
	return b
}

// Build constructs the Pipeline. Panics if no stages have been added.
func (b *PipelineBuilder) Build() *Pipeline {
	return NewPipeline(b.stages...)
}
