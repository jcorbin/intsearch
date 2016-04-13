package main

import "fmt"

type searchWatcher interface {
	beforeStep(sol *solution)
	stepped(sol *solution)
	emitted(parent, child *solution)
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

func (wat *debugWatcher) dump(sol *solution) {
	if sol.err == nil {
		fmt.Println()
		fmt.Println("Solution:")
	} else {
		fmt.Println()
		fmt.Printf("Fail: %v\n", sol.err)
	}
	trace := wat.traces[sol]
	for i, soli := range trace {
		fmt.Printf("%v %v %s\n", i, soli, soli.letterMapping())
	}
	fmt.Printf("=== %v %v\n", 0, sol)
	fmt.Printf("=== %v %s\n", 0, sol.letterMapping())
}

func (wat *debugWatcher) emitted(parent, child *solution) {
	wat.metrics.Emits++
	// TODO: not visible by searcher interface, even if wat had a reference to
	// the watched searcher
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

func (wat *debugWatcher) beforeStep(sol *solution) {
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
