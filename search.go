package main

import "fmt"

type resultFunc func(*solution)

type searcher interface {
	expand(*solution)
	step(resultFunc) bool
}

type search struct {
	frontier []*solution
	traces   map[*solution][]*solution
	debug    struct {
		expand func(sol, parent *solution)
		before func(sol *solution)
		after  func(sol *solution)
	}
	// TODO: better metric support
	metrics struct {
		Steps, Emits, MaxFrontierLen, MaxTraceLen int
	}
}

func newSearch(prob *problem) *search {
	return &search{
		frontier: make([]*solution, 0, len(prob.letterSet)),
		traces:   make(map[*solution][]*solution, len(prob.letterSet)),
	}
}

func (srch *search) dump(sol *solution) {
	if sol.err == nil {
		fmt.Println()
		fmt.Println("Solution:")
	} else {
		fmt.Println()
		fmt.Printf("Fail: %v\n", sol.err)
	}
	if srch.traces != nil {
		trace := srch.traces[sol]
		for i, soli := range trace {
			fmt.Printf("%v %v %s\n", i, soli, soli.letterMapping())
		}
	}
	fmt.Printf("=== %v %v\n", 0, sol)
	fmt.Printf("=== %v %s\n", 0, sol.letterMapping())
}

func (srch *search) expand(sol *solution) {
	srch.metrics.Emits++
	var parent *solution
	if len(srch.frontier) > 0 {
		parent = srch.frontier[0]
	}
	srch.frontier = append(srch.frontier, sol)
	if len(srch.frontier) > srch.metrics.MaxFrontierLen {
		srch.metrics.MaxFrontierLen = len(srch.frontier)
	}
	if srch.traces != nil {
		var trace []*solution
		if parent != nil {
			trace = append(trace, srch.traces[parent]...)
		}
		srch.traces[sol] = trace
		if len(trace) > srch.metrics.MaxTraceLen {
			srch.metrics.MaxTraceLen = len(trace)
		}
	}
	if srch.debug.expand != nil {
		srch.debug.expand(sol, parent)
	}
}

func (srch *search) step(result resultFunc) bool {
	for len(srch.frontier) == 0 {
		return false
	}
	sol := srch.frontier[0]
	srch.metrics.Steps++
	if srch.traces != nil {
		srch.traces[sol] = append(srch.traces[sol], sol.copy())
	}
	if srch.debug.before != nil {
		srch.debug.before(sol)
	}
	sol.step()
	if srch.debug.after != nil {
		srch.debug.after(sol)
	}

	if sol.done {
		srch.frontier = srch.frontier[1:]
		result(sol)
		if srch.traces != nil {
			delete(srch.traces, sol)
		}
	}
	return true
}

func runSearch(srch searcher, maxSteps int, init func(func(*solution)), result resultFunc) bool {
	counter := 0
	init(srch.expand)
	for srch.step(result) {
		counter++
		if counter > maxSteps {
			return false
		}
	}
	return true
}
