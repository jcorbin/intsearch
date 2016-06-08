package runnable

// EmitFunc is a state-sink function that should will the passed solution for
// eventual exploration.
type EmitFunc func(*Solution)

// ResultFunc is a terminal state processing function.  If the result function
// retains a reference to the solution, it should return true; otherwise the
// solution object will be re-used after the result function finishes.
type ResultFunc func(*Solution) bool

// InitFunc is a search initialization function.  It should call the passed
// emit function at least once to prime the frontier.  It is expected to return
// the maximum number of steps allowed before a search run is terminated on
// grounds of insanity.
type InitFunc func(EmitFunc) int

// Searcher is the interface implemented by any solution search engine.  It's
// primary use case is in relation to the SearchWatcher.
type Searcher interface {
	FrontierSize() int
	Current() *Solution
	Expand(*Solution)
	Run(init InitFunc, result ResultFunc, watcher SearchWatcher) bool
}

// Search is (the?) simplest Searcher impementation:
// - it executes one solution at a time
// - it applies no prioritization to its frontier
// - it affords no ex-post pruning of its frontier
type Search struct {
	frontier []*Solution
}

// FrontierSize returns the size of the search frontier; the number of deferred
// partial solutions waiting for eventual exploration.
func (srch *Search) FrontierSize() int {
	return len(srch.frontier)
}

// Current returns the current solution being explored, or nil if the frontier
// is empty.
func (srch *Search) Current() *Solution {
	if len(srch.frontier) > 0 {
		return srch.frontier[0]
	}
	return nil
}

// Expand adds a solution to the frontier.
func (srch *Search) Expand(sol *Solution) {
	srch.frontier = append(srch.frontier, sol)
}

// Run starts a new search run:
// - init is called to populate the frontier
// - while there is a current solution, it is stepped until it terminates
// - terminated solutions are passed to result
// - if watcher is non-nil, then its methods will be called to observe the
//   search run
func (srch *Search) Run(init InitFunc, result ResultFunc, watcher SearchWatcher) bool {
	run := searchRun{
		Search:  *srch,
		init:    init,
		result:  result,
		counter: 0,
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
	Search
	init     InitFunc
	result   ResultFunc
	maxSteps int
	counter  int
}

func (srch *searchRun) run() bool {
	srch.maxSteps = srch.init(srch.Expand)
	for sol := srch.Current(); sol != nil; sol = srch.Current() {
		for sol.Step() {
			srch.counter++
			if srch.counter > srch.maxSteps {
				return false
			}
		}
		srch.frontier = srch.frontier[1:]
		if !srch.result(sol) {
			sol.pool.Put(sol)
		}
	}
	return true
}

type searchRunWatch struct {
	searchRun
	watcher SearchWatcher
}

func (srch *searchRunWatch) expand(sol *Solution) {
	srch.watcher.Emitted(&srch.Search, sol)
	srch.searchRun.Expand(sol)
}

func (srch *searchRunWatch) run() bool {
	srch.maxSteps = srch.init(srch.expand)
	for sol := srch.Current(); sol != nil; sol = srch.Current() {
		srch.watcher.BeforeStep(&srch.Search, sol)
		if !sol.Step() {
			srch.frontier = srch.frontier[1:]
			if !srch.result(sol) {
				sol.pool.Put(sol)
			}
		}
		srch.watcher.Stepped(&srch.Search, sol)
		srch.counter++
		if srch.counter > srch.maxSteps {
			return false
		}
	}
	return true
}
