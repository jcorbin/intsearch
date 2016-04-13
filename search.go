package main

type resultFunc func(*solution)

type searcher interface {
	expand(*solution)
	step(resultFunc) bool
}

type search struct {
	frontier []*solution
	watcher  searchWatcher
}

func newSearch(prob *problem) *search {
	srch := &search{
		frontier: make([]*solution, 0, len(prob.letterSet)),
	}
	return srch
}

func (srch *search) expand(sol *solution) {
	var parent *solution
	if len(srch.frontier) > 0 {
		parent = srch.frontier[0]
	}
	srch.frontier = append(srch.frontier, sol)
	if srch.watcher != nil {
		srch.watcher.emitted(parent, sol)
	}
}

func (srch *search) step(result resultFunc) bool {
	for len(srch.frontier) == 0 {
		return false
	}
	sol := srch.frontier[0]
	if srch.watcher != nil {
		srch.watcher.beforeStep(sol)
		sol.step()
		if sol.done {
			srch.frontier = srch.frontier[1:]
			result(sol)
		}
		srch.watcher.stepped(sol)
	} else {
		sol.step()
		if sol.done {
			srch.frontier = srch.frontier[1:]
			result(sol)
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
