package main

type emitFunc func(*solution)
type resultFunc func(*solution)
type initFunc func(emitFunc)

type searcher interface {
	frontierSize() int
	current() *solution
	expand(*solution)
	run(maxSteps int, init initFunc, result resultFunc, watcher searchWatcher) bool
}

type search struct {
	frontier []*solution
	watcher  searchWatcher
}

func (srch *search) frontierSize() int {
	return len(srch.frontier)
}

func (srch *search) current() *solution {
	if len(srch.frontier) > 0 {
		return srch.frontier[0]
	}
	return nil
}

func (srch *search) expand(sol *solution) {
	if srch.watcher != nil {
		srch.watcher.emitted(srch, sol)
	}
	srch.frontier = append(srch.frontier, sol)
}

func (srch *search) run(maxSteps int, init initFunc, result resultFunc, watcher searchWatcher) bool {
	srch.watcher = watcher
	counter := 0
	init(srch.expand)

	for len(srch.frontier) > 0 {
		sol := srch.frontier[0]
		if srch.watcher != nil {
			srch.watcher.beforeStep(srch, sol)
			if !sol.step() {
				srch.frontier = srch.frontier[1:]
				result(sol)
				sol.pool.Put(sol)
			}
			srch.watcher.stepped(srch, sol)
		} else {
			if !sol.step() {
				srch.frontier = srch.frontier[1:]
				result(sol)
				sol.pool.Put(sol)
			}
		}

		counter++
		if counter > maxSteps {
			return false
		}
	}
	return true
}
