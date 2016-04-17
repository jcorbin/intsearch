package main

type emitFunc func(*solution)
type resultFunc func(*solution)
type initFunc func(emitFunc)

type searcher interface {
	frontierSize() int
	expand(*solution)
	run(maxSteps int, init initFunc, result resultFunc) bool
}

type search struct {
	frontier []*solution
	watcher  searchWatcher
}

func (srch *search) frontierSize() int {
	return len(srch.frontier)
}

func (srch *search) frontierCap() int {
	return cap(srch.frontier)
}

func (srch *search) hintFrontier(n int) {
	if cap(srch.frontier) < n {
		frontier := make([]*solution, 0, n)
		if len(srch.frontier) > 0 {
			copy(frontier, srch.frontier)
		}
		srch.frontier = frontier
	}
}

func (srch *search) expand(sol *solution) {
	var parent *solution
	if len(srch.frontier) > 0 {
		parent = srch.frontier[0]
	}
	srch.frontier = append(srch.frontier, sol)
	if srch.watcher != nil {
		srch.watcher.emitted(srch, parent, sol)
	}
}

func (srch *search) step(result resultFunc) bool {
	for len(srch.frontier) == 0 {
		return false
	}
	sol := srch.frontier[0]
	if srch.watcher != nil {
		srch.watcher.beforeStep(srch, sol)
		sol.step()
		if sol.done {
			srch.frontier = srch.frontier[1:]
			result(sol)
		}
		srch.watcher.stepped(srch, sol)
	} else {
		sol.step()
		if sol.done {
			srch.frontier = srch.frontier[1:]
			result(sol)
		}
	}
	return true
}

func (srch *search) run(maxSteps int, init initFunc, result resultFunc) bool {
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
