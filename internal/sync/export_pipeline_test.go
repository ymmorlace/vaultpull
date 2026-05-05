package sync

// ExportStage exposes the Stage type for black-box tests that need to
// construct stages without importing internal symbols directly.
type ExportStage = Stage

// ExportNewPipeline re-exports NewPipeline for use in external test packages.
var ExportNewPipeline = NewPipeline

// ExportNewPipelineBuilder re-exports NewPipelineBuilder.
var ExportNewPipelineBuilder = NewPipelineBuilder
