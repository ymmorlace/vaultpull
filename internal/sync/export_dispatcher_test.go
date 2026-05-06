package sync

// ExportJob re-exports Job for use in external test packages.
type ExportJob = Job

// ExportDispatchResult re-exports DispatchResult for use in external test packages.
type ExportDispatchResult = DispatchResult

// NewDispatcherExported wraps NewDispatcher for external test access.
func NewDispatcherExported(writer EnvWriter, workers int) *Dispatcher {
	return NewDispatcher(writer, workers)
}
