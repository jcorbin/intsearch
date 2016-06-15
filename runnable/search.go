package runnable

import "github.com/jcorbin/intsearch/word"

// SearchPlan implements word.Plan by running the generated steps.
type SearchPlan struct {
	*StepGen
}

// Run runs the generated steps.
func (sp *SearchPlan) Run(res word.Resultor) {
	run := searchRun{
		// large upper limit for the search execution: run every step for every
		// possible brute force solution
		maxSteps: fallFact(sp.Base, len(sp.Letters)) * len(sp.steps),
		result:   res,
		counter:  0,
	}
	run.expand(newSolution(&sp.PlanProblem.Problem, sp.steps, run.expand))
	run.run()
}

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
	result   word.Resultor
	maxSteps int
	counter  int
}

func (srch *searchRun) run() bool {
	for sol := srch.current(); sol != nil; sol = srch.current() {
		for sol.Step() {
			srch.counter++
			if srch.counter > srch.maxSteps {
				return false
			}
		}
		srch.frontier = srch.frontier[1:]
		if srch.result.Result(sol) {
			break
		}
		sol.pool.Put(sol)
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
