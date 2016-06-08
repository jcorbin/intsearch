package runnable

// TraceWatcher implements a solution execution trace collector.
type TraceWatcher struct{}

// NewTraceWatcher creates a new trace watcher.
func NewTraceWatcher() *TraceWatcher {
	return &TraceWatcher{}
}

// Emitted shallow copies any parent trace to the new child.
func (tw *TraceWatcher) Emitted(srch Searcher, child *Solution) {
	if parent := srch.Current(); parent != nil {
		child.trace = append([]*Solution(nil), parent.trace...)
	} else {
		child.trace = nil
	}
	// TODO: want?
	// if len(trace) > wat.metrics.MaxTraceLen {
	// 	wat.metrics.MaxTraceLen = len(trace)
	// }
}

// BeforeStep appends a copy of the solution to its trace.
func (tw *TraceWatcher) BeforeStep(srch Searcher, sol *Solution) {
	sol.trace = append(sol.trace, sol.copy())
}

// Stepped does nothing.
func (tw *TraceWatcher) Stepped(srch Searcher, sol *Solution) {
}
