package runnable

import "github.com/jcorbin/intsearch/word"

// SearchPlan implements word.Plan by running the generated steps.
type SearchPlan struct {
	*StepGen
}

// Run runs the generated steps.
func (sp *SearchPlan) Run(res word.Resultor) {
	var srch Search
	srch.Run(
		func(emit EmitFunc) int {
			emit(newSolution(&sp.PlanProblem.Problem, sp.steps, emit))
			// worst case, we have to run every step for every possible brute force solution
			numBrute := fallFact(sp.Base, len(sp.Letters))
			return numBrute * len(sp.steps)
		},
		func(sol *Solution) bool {
			res.Result(sol)
			return false
		})
}

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

// Search is a simple serial searcher.
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
func (srch *Search) Run(init InitFunc, result ResultFunc) bool {
	run := searchRun{
		Search:  *srch,
		init:    init,
		result:  result,
		counter: 0,
	}
	return run.run()
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

func fallFact(x, y int) int {
	z := 1
	for y > 0 {
		z *= x
		x--
		y--
	}
	return z
}
