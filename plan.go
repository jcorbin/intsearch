package main

import (
	"fmt"
	"log"
	"strings"
)

type carryValue int

type planFunc func(*planProblem, solutionGen, bool)

func planNaiveBrute(prob *planProblem, gen solutionGen, verified bool) {
	gen.init("naive brute force")
	for _, c := range prob.sortedLetters() {
		prob.chooseRange(gen, c, 0, prob.base-1)
	}
	gen.check(errCheckFailed)
	if verified {
		gen.verify()
	}
	gen.finish()
	gen.finalize()
}

func planPrunedBrute(prob *planProblem, gen solutionGen, verified bool) {
	gen.init("pruned brute force")
	var mins [256]int
	for _, word := range prob.words {
		mins[word[0]] = 1
	}
	for i := len(prob.columns) - 1; i >= 0; i-- {
		col := &prob.columns[i]
		for _, c := range col.cx {
			if c != 0 && !prob.known[c] {
				prob.chooseRange(gen, c, mins[c], prob.base-1)
			}
		}
		prob.checkColumn(gen, col)
	}
	if verified {
		gen.verify()
	}
	gen.finish()
	gen.finalize()
}

func planTopDown(prob *planProblem, gen solutionGen, verified bool) {
	gen.init("top down ... bottom up")
	if !prob.procTopDown(gen, &prob.columns[0], verified) {
		panic("unable to plan top down")
	}
	gen.finalize()
}

func planBottomUp(prob *planProblem, gen solutionGen, verified bool) {
	gen.init("bottom up")
	prob.procBottomUp(gen, verified)
	gen.finalize()
}

const (
	carryUnknown carryValue = iota - 1
	carryZero
	carryOne
	carryComputed
)

//go:generate stringer -type=carryValue

func (cv carryValue) expr() string {
	switch cv {
	case carryUnknown:
		return "?"
	case carryZero:
		return "0"
	case carryOne:
		return "1"
	case carryComputed:
		return "C"
	default:
		return "!"
	}
}

type column struct {
	i       int
	prior   *column
	cx      [3]byte
	solved  bool
	have    int
	known   int
	unknown int
	fixed   int
	carry   carryValue
}

func (col *column) String() string {
	return fmt.Sprintf(
		"%s solved=%t have=%d known=%d unknown=%d fixed=%d",
		col.label(),
		col.solved, col.have, col.known, col.unknown, col.fixed)
}

func (col *column) label() string {
	return fmt.Sprintf("[%d] %s carry=%s", col.i, col.expr(), col.carry.expr())
}

func (col *column) expr() string {
	parts := make([]string, 0, 7)
	if col.prior != nil {
		parts = append(parts, col.prior.carry.expr())
	}
	for _, c := range col.cx[:2] {
		if c != 0 {
			if len(parts) > 0 {
				parts = append(parts, "+", string(c))
			} else {
				parts = append(parts, string(c))
			}
		}
	}
	parts = append(parts, "=", string(col.cx[2]))
	return strings.Join(parts, " ")
}

func (col *column) priorCarry() carryValue {
	if col.prior == nil {
		return carryZero
	}
	return col.prior.carry
}

type planProblem struct {
	problem
	columns      []column
	letCols      map[byte][]*column
	known        map[byte]bool
	fixedLetters map[byte]int
	fixedValues  []bool
}

type solutionGen interface {
	logf(string, ...interface{}) error
	init(desc string)
	fix(c byte, v int)
	computeSum(col *column)
	computeFirstSummand(col *column)
	computeSecondSummand(col *column)
	chooseRange(c byte, min, max int)
	checkColumn(col *column, err error)
	check(err error)
	finish()
	verify()
	finalize()
}

func newPlanProblem(p *problem) *planProblem {
	C := p.numColumns()
	N := len(p.letterSet)
	prob := &planProblem{
		problem:      *p,
		columns:      make([]column, C),
		letCols:      make(map[byte][]*column, N),
		known:        make(map[byte]bool, N),
		fixedLetters: make(map[byte]int, N),
		fixedValues:  make([]bool, p.base),
	}
	var last *column
	for i := 0; i < C; i++ {
		col := &prob.columns[i]
		col.i = i
		if i == 0 {
			col.carry = carryZero
		} else {
			col.carry = carryUnknown
		}
		if last != nil {
			last.prior = col
		}
		col.cx = prob.getColumn(i)
		a, b, c := col.cx[0], col.cx[1], col.cx[2]
		if a != 0 {
			col.have++
			col.unknown++
			prob.letCols[a] = append(prob.letCols[a], col)
		}
		if b != 0 {
			col.have++
			if b != a {
				col.unknown++
				prob.letCols[b] = append(prob.letCols[b], col)
			}
		}
		if c != 0 {
			col.have++
			if c != b && c != a {
				col.unknown++
				prob.letCols[c] = append(prob.letCols[c], col)
			}
		}
		last = col
	}
	return prob
}

func (prob *planProblem) markKnown(c byte) {
	prob.known[c] = true
	for _, col := range prob.letCols[c] {
		col.unknown--
		col.known++
	}
}

func (prob *planProblem) procTopDown(gen solutionGen, col *column, verified bool) bool {
	if col.prior == nil {
		prob.solveColumn(gen, col)
		if verified {
			gen.verify()
		}
		gen.finish()
		return true
	}

	if prob.maySolveColumn(gen, col) {
		return prob.procTopDown(gen, col.prior, verified)
	}

	prob.procBottomUp(gen, verified)
	return true
}

func (prob *planProblem) procBottomUp(gen solutionGen, verified bool) {
	for i := len(prob.columns) - 1; i >= 0; i-- {
		prob.solveColumn(gen, &prob.columns[i])
	}
	if verified {
		gen.verify()
	}
	gen.finish()
}

func (prob *planProblem) checkColumn(gen solutionGen, col *column) bool {
	if !col.solved {
		gen.checkColumn(col, nil)
		col.solved = true
		col.carry = carryComputed
	}
	return true
}

func (prob *planProblem) maySolveColumn(gen solutionGen, col *column) bool {
	if col.solved {
		if col.unknown != 0 {
			panic("invalid column solved state")
		}
		return true
	}

	if col.unknown == 0 {
		return prob.checkColumn(gen, col)
	}

	if col.have == 1 {
		if prob.solveSingularColumn(gen, col) {
			return true
		}
	}

	return prob.solveColumnFromPrior(gen, col)
}

func (prob *planProblem) solveColumn(gen solutionGen, col *column) {
	if !prob.maySolveColumn(gen, col) {
		log.Fatalf("cannot solve column: %#v", col)
	}
}

func (prob *planProblem) fix(gen solutionGen, c byte, v int) {
	prob.fixedLetters[c] = v
	prob.fixedValues[v] = true
	// TODO: consider inlining markKnown and unifying the for range letCols loops
	prob.markKnown(c)
	for _, col := range prob.letCols[c] {
		col.fixed++
	}
	gen.fix(c, v)
}

func (prob *planProblem) solveSingularColumn(gen solutionGen, col *column) bool {
	if col.have != 1 || col.unknown != 1 {
		return false
	}

	if c := col.cx[2]; col.i == 0 && c != 0 {
		// carry + _ + _ = c --> c == carry --> c = carry = 1
		if col.prior == nil {
			panic("invalid final column: has no prior")
		}
		prob.fix(gen, c, 1)
		col.solved = true
		col.prior.carry = carryOne
		return true
	}

	return false
}

func (prob *planProblem) fixRange(min, max int, c byte) (int, int) {
	if min == 0 && (prob.fixedValues[0] ||
		c == prob.words[0][0] ||
		c == prob.words[1][0] ||
		c == prob.words[2][0]) {
		min = 1
	}
	for max > 0 && prob.fixedValues[max] {
		max--
	}
	for min <= max && prob.fixedValues[min] {
		min++
	}
	if min > max {
		panic("no choices possible")
	}
	return min, max
}

func (prob *planProblem) chooseOne(gen solutionGen, col *column) bool {
	return prob.chooseFirst(gen, col)
}

func (prob *planProblem) chooseFirst(gen solutionGen, col *column) bool {
	for _, cc := range col.cx {
		if cc == 0 || prob.known[cc] {
			continue
		}
		min, max := prob.fixRange(0, prob.base-1, cc)
		if min == max {
			prob.fix(gen, cc, min)
			return true
		}
		prob.chooseRange(gen, cc, min, max)
		return true
	}
	return false
}

func (prob *planProblem) chooseBest(gen solutionGen, col *column) bool {
	var min, max, N [3]int

	i := -1
	for j, cc := range col.cx {
		if cc == 0 || prob.known[cc] {
			continue
		}
		min[j], max[j] = prob.fixRange(0, prob.base-1, cc)
		if min[j] == max[j] {
			prob.fix(gen, cc, min[j])
			return true
		}
		N[j] = max[j] - min[j]
		if i < 0 || N[j] < N[i] {
			i = j
		}
	}
	if i < 0 {
		return false
	}

	prob.chooseRange(gen, col.cx[i], min[i], max[i])
	return true
}

func (prob *planProblem) chooseRange(gen solutionGen, c byte, min, max int) {
	gen.chooseRange(c, min, max)
	prob.markKnown(c)
}

func (prob *planProblem) solveColumnFromPrior(gen solutionGen, col *column) bool {
	if col.priorCarry() == carryUnknown {
		// unknown prior carry not yet support; i.e. solveColumn must be called
		// in bottom-up/right-to-left/decreasing-index order
		return false
	}

	if col.unknown == 0 {
		// this is checkColumn's job
		return false
	}

	gen.logf("solveFromPrior: %s", col.label())

	for u := col.unknown; u > 1; u = col.unknown {
		if !prob.chooseOne(gen, col) {
			break
		}
	}
	if col.unknown > 1 {
		// chooseOne was unable to figure it out
		return false
	}

	if c := col.cx[0]; c != 0 && !prob.known[c] {
		gen.computeFirstSummand(col)
		col.carry = carryComputed
		prob.markKnown(c)
	} else if c := col.cx[1]; c != 0 && !prob.known[c] {
		gen.computeSecondSummand(col)
		col.carry = carryComputed
		prob.markKnown(c)
	} else if c := col.cx[2]; c != 0 && !prob.known[c] {
		gen.computeSum(col)
		col.carry = carryComputed
		prob.markKnown(c)
	} else {
		panic("invalid solveColumnFromPrior state")
	}

	col.solved = true
	return true
}
