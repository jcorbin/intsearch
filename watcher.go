package main

import "fmt"

type searchWatcher interface {
	beforeStep(srch searcher, sol *solution)
	stepped(srch searcher, sol *solution)
	emitted(srch searcher, parent, child *solution)
}

type debugWatcher struct {
	traces map[*solution][]*solution
	debug  struct {
		expand func(sol, parent *solution)
		before func(sol *solution)
		after  func(sol *solution)
	}
	// TODO: better metric support
	metrics struct {
		Steps, Emits, MaxFrontierLen, MaxTraceLen int
	}
}

func newDebugWatcher(prob *problem) *debugWatcher {
	return &debugWatcher{
		traces: make(map[*solution][]*solution, len(prob.letterSet)),
	}
}

func (wat *debugWatcher) dump(srch searcher, sol *solution) {
	sol.dump()
	trace := wat.traces[sol]
	for i, soli := range trace {
		fmt.Printf("%v %v %s\n", i, soli, soli.letterMapping())
	}
}

func (wat *debugWatcher) emitted(srch searcher, parent, child *solution) {
	wat.metrics.Emits++
	// TODO: not visible by searcher interface
	// if len(srch.frontier) > wat.metrics.MaxFrontierLen {
	// 	wat.metrics.MaxFrontierLen = len(srch.frontier)
	// }
	var trace []*solution
	if parent != nil {
		trace = append(trace, wat.traces[parent]...)
	}
	wat.traces[child] = trace
	if len(trace) > wat.metrics.MaxTraceLen {
		wat.metrics.MaxTraceLen = len(trace)
	}
	if wat.debug.expand != nil {
		wat.debug.expand(child, parent)
	}
}

func (wat *debugWatcher) beforeStep(srch searcher, sol *solution) {
	wat.metrics.Steps++
	wat.traces[sol] = append(wat.traces[sol], sol.copy())
	if wat.debug.before != nil {
		wat.debug.before(sol)
	}
}

func (wat *debugWatcher) stepped(sol *solution) {
	if wat.debug.after != nil {
		wat.debug.after(sol)
	}
	if sol.done {
		delete(wat.traces, sol)
	}
}
