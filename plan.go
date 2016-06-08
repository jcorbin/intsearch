package main

import (
	"log"

	"github.com/jcorbin/intsearch/word"
)

type planFunc func(*word.PlanProblem, word.SolutionGen, bool)

func planNaiveBrute(prob *word.PlanProblem, gen word.SolutionGen, verified bool) {
	gen.Init("naive brute force")
	for _, c := range prob.SortedLetters() {
		prob.ChooseRange(gen, c, 0, prob.Base-1)
	}
	gen.Check(errCheckFailed)
	if verified {
		gen.Verify()
	}
	gen.Finish()
	gen.Finalize()
}

func planPrunedBrute(prob *word.PlanProblem, gen word.SolutionGen, verified bool) {
	gen.Init("pruned brute force")
	var mins [256]int
	for _, word := range prob.Words {
		mins[word[0]] = 1
	}
	for i := len(prob.Columns) - 1; i >= 0; i-- {
		col := &prob.Columns[i]
		for _, c := range col.Chars {
			if c != 0 && !prob.Known[c] {
				prob.ChooseRange(gen, c, mins[c], prob.Base-1)
			}
		}
		prob.CheckColumn(gen, col)
	}
	if verified {
		gen.Verify()
	}
	gen.Finish()
	gen.Finalize()
}

func planTopDown(prob *word.PlanProblem, gen word.SolutionGen, verified bool) {
	var proc func(prob *word.PlanProblem, gen word.SolutionGen, col *word.Column) bool
	proc = func(prob *word.PlanProblem, gen word.SolutionGen, col *word.Column) bool {
		if col.Prior == nil {
			if !prob.MaySolveColumn(gen, col) {
				return false
			}
			if verified {
				gen.Verify()
			}
			gen.Finish()
			return true
		}
		if prob.MaySolveColumn(gen, col) {
			return proc(prob, gen, col.Prior)
		}
		return prob.AssumeCarrySolveColumn(gen, col, proc)
	}
	gen.Init("top down")
	if !proc(prob, gen, &prob.Columns[0]) {
		panic("unable to plan top down")
	}
	gen.Finalize()
}

func planBottomUp(prob *word.PlanProblem, gen word.SolutionGen, verified bool) {
	gen.Init("bottom up")
	for i := len(prob.Columns) - 1; i >= 0; i-- {
		col := &prob.Columns[i]
		if !prob.MaySolveColumn(gen, col) {
			log.Fatalf("cannot solve column: %#v", col)
		}
	}
	if verified {
		gen.Verify()
	}
	gen.Finish()
	gen.Finalize()
}
