package runnable

import "github.com/jcorbin/intsearch/word"

// SearchPlan implements word.Plan by running the generated steps.
type SearchPlan struct {
	*StepGen
}

// Run runs the generated steps.
func (sp *SearchPlan) Run(res word.Resultor) {
	run := searchRun{
		init: func(emit EmitFunc) int {
			emit(newSolution(&sp.PlanProblem.Problem, sp.steps, emit))
			// worst case, we have to run every step for every possible brute force solution
			numBrute := fallFact(sp.Base, len(sp.Letters))
			return numBrute * len(sp.steps)
		},
		result: func(sol *Solution) bool {
			res.Result(sol)
			return false
		},
		counter: 0,
	}
	run.run()
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

type search struct {
	frontier []*Solution
}

func (srch *search) current() *Solution {
	if len(srch.frontier) > 0 {
		return srch.frontier[0]
	}
	return nil
}

func (srch *search) expand(sol *Solution) {
	srch.frontier = append(srch.frontier, sol)
}

type searchRun struct {
	search
	init     InitFunc
	result   ResultFunc
	maxSteps int
	counter  int
}

func (srch *searchRun) run() bool {
	srch.maxSteps = srch.init(srch.expand)
	for sol := srch.current(); sol != nil; sol = srch.current() {
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
