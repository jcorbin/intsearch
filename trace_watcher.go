package main

import "fmt"

type traceWatcher map[*solution][]*solution

func newTraceWatcher(prob *problem) traceWatcher {
	return traceWatcher(make(map[*solution][]*solution, len(prob.letterSet)))
}

func (traces traceWatcher) dump(sol *solution) {
	trace := traces[sol]
	for i, soli := range trace {
		fmt.Printf("%v %v %s\n", i, soli, soli.letterMapping())
	}
}

func (traces traceWatcher) emitted(srch searcher, parent, child *solution) {
	var trace []*solution
	if parent != nil {
		trace = append(trace, traces[parent]...)
	}
	traces[child] = trace
	// TODO: want?
	// if len(trace) > wat.metrics.MaxTraceLen {
	// 	wat.metrics.MaxTraceLen = len(trace)
	// }
}

func (traces traceWatcher) beforeStep(srch searcher, sol *solution) {
	traces[sol] = append(traces[sol], sol.copy())
}

func (traces traceWatcher) stepped(srch searcher, sol *solution) {
	if sol.done {
		delete(traces, sol)
	}
}
