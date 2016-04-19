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
	srch.frontier = append(srch.frontier, sol)
}

func (srch *search) run(maxSteps int, init initFunc, result resultFunc, watcher searchWatcher) bool {
	run := searchRun{
		search:   *srch,
		init:     init,
		result:   result,
		maxSteps: maxSteps,
		counter:  0,
	}

	if watcher == nil {
		return run.run()
	}

	watrun := searchRunWatch{
		searchRun: run,
		watcher:   watcher,
	}

	return watrun.run()
}

type searchRun struct {
	search
	init     initFunc
	result   resultFunc
	maxSteps int
	counter  int
}

func (srch *searchRun) run() bool {
	srch.init(srch.expand)
	for sol := srch.current(); sol != nil; sol = srch.current() {
		for sol.step() {
			srch.counter++
			if srch.counter > srch.maxSteps {
				return false
			}
		}
		srch.frontier = srch.frontier[1:]
		srch.result(sol)
		sol.pool.Put(sol)
	}
	return true
}

type searchRunWatch struct {
	searchRun
	watcher searchWatcher
}

func (srch *searchRunWatch) expand(sol *solution) {
	srch.watcher.emitted(&srch.search, sol)
	srch.searchRun.expand(sol)
}

func (srch *searchRunWatch) run() bool {
	srch.init(srch.expand)
	for sol := srch.current(); sol != nil; sol = srch.current() {
		srch.watcher.beforeStep(&srch.search, sol)
		if !sol.step() {
			srch.frontier = srch.frontier[1:]
			srch.result(sol)
			sol.pool.Put(sol)
		}
		srch.watcher.stepped(&srch.search, sol)
		srch.counter++
		if srch.counter > srch.maxSteps {
			return false
		}
	}
	return true
}
