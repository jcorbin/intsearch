package main

import "fmt"

type search struct {
	frontier []*solution
	traces   map[*solution][]*solution
	init     func(func(*solution))
	debug    struct {
		before func(sol *solution)
		after  func(sol *solution)
	}
	result func(*solution, []*solution)
	// TODO: better metric support
	metrics struct {
		Steps, Emits, MaxFrontierLen, MaxTraceLen int
	}
}

func (srch *search) dump(sol *solution, trace []*solution) {
	if sol.err == nil {
		fmt.Println()
		fmt.Println("Solution:")
	} else {
		fmt.Println()
		fmt.Printf("Fail: %v\n", sol.err)
	}
	for i, soli := range trace {
		fmt.Printf("%v %v %s\n", i, soli, soli.letterMapping())
	}
	fmt.Printf("=== %v %v\n", 0, sol)
	fmt.Printf("=== %v %s\n", 0, sol.letterMapping())
}

func (srch *search) emit(sol *solution) {
	// fmt.Printf("+++ %v %v\n", len(srch.frontier), sol)
	srch.metrics.Emits++
	srch.frontier = append(srch.frontier, sol)
	if len(srch.frontier) > srch.metrics.MaxFrontierLen {
		srch.metrics.MaxFrontierLen = len(srch.frontier)
	}
	parent := srch.frontier[0]
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
}

func (srch *search) step(sol *solution) {
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
		var trace []*solution
		if srch.traces != nil {
			trace = srch.traces[sol]
			delete(srch.traces, sol)
		}
		srch.result(sol, trace)
	}
}

func (srch *search) run(maxSteps int) bool {
	srch.init(srch.emit)
	counter := 0
	for len(srch.frontier) > 0 {
		counter++
		if counter > maxSteps {
			return false
		}
		srch.step(srch.frontier[0])
	}
	return true
}
