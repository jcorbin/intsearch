package main

type traceWatcher struct{}

func newTraceWatcher() *traceWatcher {
	return &traceWatcher{}
}

func (tw *traceWatcher) emitted(srch searcher, child *solution) {
	if parent := srch.current(); parent != nil {
		child.trace = append([]*solution(nil), parent.trace...)
	} else {
		child.trace = nil
	}
	// TODO: want?
	// if len(trace) > wat.metrics.MaxTraceLen {
	// 	wat.metrics.MaxTraceLen = len(trace)
	// }
}

func (tw *traceWatcher) beforeStep(srch searcher, sol *solution) {
	sol.trace = append(sol.trace, sol.copy())
}

func (tw *traceWatcher) stepped(srch searcher, sol *solution) {
}
