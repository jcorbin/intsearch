package word

import (
	"errors"
	"log"
)

var errBruteCheckFailed = errors.New("check failed")

// Plan is a concrete plan that can be ran to find one or more solutions to a
// problem.
type Plan interface {
	Dump(logf func(format string, args ...interface{}))
	Decorate(args ...interface{}) []string
	Run(Resultor)
}

// PlanFunc is the type of a concrete solution strategy.  The function should:
// - call gen.Init
// - use some combinatino of gen method calls to solve every column, or at
//   least determine every letter
// - if verified is true, call gen.Verify (also call .Verify on any Fork'd
//   alternates)
// - call gen.Finish (also call .Finish on any Fork'd alternates)
// - call gen.Finalize
type PlanFunc func(SolutionGen, bool) Plan

// PlanNaiveBrute implements the most naive (factorial in the number of letters
// branching factor) strategy:
// - gen.ChooseRange for every letter (in naive sorted order)
// - gen.Check to filter
func PlanNaiveBrute(gen SolutionGen, verified bool) Plan {
	gen.Init("naive brute force")
	prob := gen.Problem()
	for _, c := range prob.SortedLetters() {
		prob.ChooseRange(gen, c, 0, prob.Base-1)
	}
	gen.Check(errBruteCheckFailed)
	if verified {
		gen.Verify()
	}
	gen.Finish()
	return gen.Finalize()
}

// PlanPrunedBrute implements a sane brute force search:
// - for each column right to left ...
// - ... choose all three letters
// - ... then check the column (pruning any choices that weren't valid for this column)
//
// This strategy wins over naive brute since there is a drastic search space
// reduction as we move from column to column.
func PlanPrunedBrute(gen SolutionGen, verified bool) Plan {
	gen.Init("pruned brute force")
	prob := gen.Problem()
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
	return gen.Finalize()
}

// PlanBottomUp implements a right-to-left (bottom-up) solver:
// - for each column from the right ...
// - ... choose letters until only one is unknown
// - ... compute the remaining unknown letter
//
// This strategy wins over pruned brute because it avoids many high branching
// factor choices, instead performing a direct computation when possible.
func PlanBottomUp(gen SolutionGen, verified bool) Plan {
	gen.Init("bottom up")
	prob := gen.Problem()
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
	return gen.Finalize()
}

// PlanTopDown implements a left-to-right (top-down) solver:
// - for each column from the left ...
// - ... create an alternate path of execution, and assume prior column (next
//   to the right) carry to be 0 in the current path, and 1 in the alternate
//   path.
// - ... choose letters until only one is unknown
// - ... compute the remaining unknown letter
//
// This strategy wins over bottom up because it reduces branching factor in
// multiple ways:
// - the carry assumption fork is a static 2x branch, compared to the first B
//   branches of bottom up
// - furthermore initial letter choices cannot be zero, causing all subsequent
//   range choices to be reduced by 1
func PlanTopDown(gen SolutionGen, verified bool) Plan {
	prob := gen.Problem()
	var proc func(prob *PlanProblem, gen SolutionGen, col *Column) bool
	proc = func(prob *PlanProblem, gen SolutionGen, col *Column) bool {
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
	return gen.Finalize()
}
