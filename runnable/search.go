package runnable

import "github.com/jcorbin/intsearch/word"

// SearchPlan implements word.Plan by running the generated steps.
type SearchPlan struct {
	*word.Problem
	steps     []Step
	addrAnnos map[int][]string
}

// Decorate returns a list of any known annotations any Solution arguments.
func (sp *SearchPlan) Decorate(args ...interface{}) []string {
	if sp.addrAnnos == nil {
		return nil
	}
	var dec []string
	for _, arg := range args {
		if sol, ok := arg.(*Solution); ok {
			if addr := sol.stepi; addr > len(sp.steps) {
				dec = append(dec, "INVALID")
			} else if annos := sp.addrAnnos[addr]; len(annos) > 0 {
				dec = append(dec, annos...)
			}
		}
	}
	return dec
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
	run.expand(newSolution(sp.Problem, sp.steps, run.expand))
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
