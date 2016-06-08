package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/jcorbin/intsearch/word"
)

type planFunc func(*planProblem, solutionGen, bool)

func planNaiveBrute(prob *planProblem, gen solutionGen, verified bool) {
	gen.Init("naive brute force")
	for _, c := range prob.SortedLetters() {
		prob.chooseRange(gen, c, 0, prob.Base-1)
	}
	gen.Check(errCheckFailed)
	if verified {
		gen.Verify()
	}
	gen.Finish()
	gen.Finalize()
}

func planPrunedBrute(prob *planProblem, gen solutionGen, verified bool) {
	gen.Init("pruned brute force")
	var mins [256]int
	for _, word := range prob.Words {
		mins[word[0]] = 1
	}
	for i := len(prob.columns) - 1; i >= 0; i-- {
		col := &prob.columns[i]
		for _, c := range col.Chars {
			if c != 0 && !prob.known[c] {
				prob.chooseRange(gen, c, mins[c], prob.Base-1)
			}
		}
		prob.checkColumn(gen, col)
	}
	if verified {
		gen.Verify()
	}
	gen.Finish()
	gen.Finalize()
}

func planTopDown(prob *planProblem, gen solutionGen, verified bool) {
	gen.Init("top down")
	if !prob.procTopDown(gen, &prob.columns[0], verified) {
		panic("unable to plan top down")
	}
	gen.Finalize()
}

func planBottomUp(prob *planProblem, gen solutionGen, verified bool) {
	gen.Init("bottom up")
	prob.procBottomUp(gen, verified)
	gen.Finalize()
}

type planProblemPool struct {
	sync.Pool
}

func (pp *planProblemPool) Get() *planProblem {
	if item := pp.Pool.Get(); item != nil {
		return item.(*planProblem)
	}
	return nil
}

func (pp *planProblemPool) Put(prob *planProblem) {
	pp.Pool.Put(prob)
}

type planProblem struct {
	word.Problem
	pool         planProblemPool
	annotated    bool
	columns      []word.Column
	letCols      map[byte][]*word.Column
	known        map[byte]bool
	fixedLetters map[byte]int
	fixedValues  []bool
	remap        map[*word.Column]*word.Column
}

type solutionGen interface {
	Logf(string, ...interface{}) error
	Init(desc string)
	Fork(prob *planProblem, name, alt, cont string) solutionGen
	Fix(c byte, v int)
	ComputeSum(col *word.Column)
	ComputeFirstSummand(col *word.Column)
	ComputeSecondSummand(col *word.Column)
	ChooseRange(c byte, min, max int)
	CheckColumn(col *word.Column, err error)
	Check(err error)
	Finish()
	Verify()
	Finalize()
}

func newPlanProblem(p *word.Problem, annotated bool) *planProblem {
	C := p.NumColumns()
	N := len(p.Letters)
	prob := &planProblem{
		Problem:      *p,
		annotated:    annotated,
		columns:      make([]word.Column, C),
		letCols:      make(map[byte][]*word.Column, N),
		known:        make(map[byte]bool, N),
		fixedLetters: make(map[byte]int, N),
		fixedValues:  make([]bool, p.Base),
	}
	var last *word.Column
	for i := 0; i < C; i++ {
		col := &prob.columns[i]
		col.I = i
		if i == 0 {
			col.Carry = word.CarryZero
		} else {
			col.Carry = word.CarryUnknown
		}
		if last != nil {
			last.Prior = col
		}
		col.Chars = prob.GetColumn(i)
		a, b, c := col.Chars[0], col.Chars[1], col.Chars[2]
		if a != 0 {
			col.Have++
			col.Unknown++
			prob.letCols[a] = append(prob.letCols[a], col)
		}
		if b != 0 {
			col.Have++
			if b != a {
				col.Unknown++
				prob.letCols[b] = append(prob.letCols[b], col)
			}
		}
		if c != 0 {
			col.Have++
			if c != b && c != a {
				col.Unknown++
				prob.letCols[c] = append(prob.letCols[c], col)
			}
		}
		last = col
	}
	return prob
}

func (prob *planProblem) copy() *planProblem {
	C := prob.NumColumns()
	N := len(prob.Letters)

	other := prob.pool.Get()
	if other == nil {
		other = &planProblem{
			pool:      prob.pool,
			remap:     prob.remap,
			Problem:   prob.Problem,
			annotated: prob.annotated,
		}
		other.columns = make([]word.Column, C)
		other.letCols = make(map[byte][]*word.Column, N)
		other.known = make(map[byte]bool, N)
		other.fixedLetters = make(map[byte]int, N)
		other.fixedValues = append([]bool(nil), prob.fixedValues...)
	} else {
		if len(other.columns) != C {
			other.columns = make([]word.Column, C)
		}
		if len(other.fixedValues) != len(prob.fixedValues) {
			other.fixedValues = append([]bool(nil), prob.fixedValues...)
		}
		for l := range other.fixedLetters {
			delete(other.fixedLetters, l)
		}
		for c := range other.known {
			delete(other.known, c)
		}
	}

	for l, v := range prob.fixedLetters {
		other.fixedLetters[l] = v
	}

	for c, k := range prob.known {
		other.known[c] = k
	}

	if prob.remap == nil {
		prob.remap = make(map[*word.Column]*word.Column, len(prob.columns))
		other.remap = prob.remap
	}
	remap := prob.remap

	var last *word.Column
	for i := 0; i < C; i++ {
		other.columns[i] = prob.columns[i]
		col := &other.columns[i]
		remap[&prob.columns[i]] = col
		if last != nil {
			last.Prior = col
		}
		last = col
	}

	for c, cols := range prob.letCols {
		otherCols, _ := other.letCols[c]
		if otherCols == nil {
			otherCols = make([]*word.Column, len(cols))
		}
		for i, col := range cols {
			otherCols[i] = remap[col]
		}
		other.letCols[c] = otherCols
	}

	return other
}

func (prob *planProblem) markKnown(c byte) {
	prob.known[c] = true
	for _, col := range prob.letCols[c] {
		col.Unknown--
		col.Known++
	}
}

func (prob *planProblem) procTopDown(gen solutionGen, col *word.Column, verified bool) bool {
	if col.Prior == nil {
		prob.solveColumn(gen, col)
		if verified {
			gen.Verify()
		}
		gen.Finish()
		return true
	}

	if prob.maySolveColumn(gen, col) {
		return prob.procTopDown(gen, col.Prior, verified)
	}

	return prob.assumeCarrySolveColumn(
		gen, col,
		func(subProb *planProblem, subGen solutionGen, subCol *word.Column) bool {
			return subProb.procTopDown(subGen, subCol, verified)
		})
}

func (prob *planProblem) assumeCarrySolveColumn(
	gen solutionGen, col *word.Column,
	andThen func(*planProblem, solutionGen, *word.Column) bool,
) bool {
	var label, altLabel, contLabel string
	if prob.annotated {
		label = col.Label()
		gen.Logf("assumeCarrySolveColumn: %s", label)
		label = fmt.Sprintf("assumeCarry(%s)", label)
	}

	altProb := prob.copy()
	altCol := &altProb.columns[col.I]

	altCol.Prior.Carry = word.CarryZero
	col.Prior.Carry = word.CarryOne
	if prob.annotated {
		altLabel = fmt.Sprintf("assumeCarry(%s)", altCol.Label())
		contLabel = fmt.Sprintf("assumeCarry(%s)", col.Label())
	}
	altGen := gen.Fork(altProb, label, altLabel, contLabel)

	altProb.solveColumn(altGen, altCol)
	if !andThen(altProb, altGen, altCol.Prior) {
		// TODO: needs to be able to cancel the fork, leaving only the cont
		// path below
		panic("alt pruning unimplemented")
	}
	prob.pool.Put(altProb)
	altProb, altGen, altCol = nil, nil, nil

	prob.solveColumn(gen, col)
	if !andThen(prob, gen, col.Prior) {
		// TODO: needs to be able replace the fork with just the generated alt
		// steps, or return false if the alt failed as well
		panic("alt swapping unimplemented")
	}

	return true
}

func (prob *planProblem) procBottomUp(gen solutionGen, verified bool) {
	for i := len(prob.columns) - 1; i >= 0; i-- {
		prob.solveColumn(gen, &prob.columns[i])
	}
	if verified {
		gen.Verify()
	}
	gen.Finish()
}

func (prob *planProblem) checkColumn(gen solutionGen, col *word.Column) bool {
	if !col.Solved {
		gen.CheckColumn(col, nil)
		col.Solved = true
		col.Carry = word.CarryComputed
	}
	return true
}

func (prob *planProblem) maySolveColumn(gen solutionGen, col *word.Column) bool {
	if col.Solved {
		if col.Unknown != 0 {
			panic("invalid column solved state")
		}
		return true
	}

	if col.Unknown == 0 {
		return prob.checkColumn(gen, col)
	}

	if col.Have == 1 {
		if prob.solveSingularColumn(gen, col) {
			return true
		}
	}

	return prob.solveColumnFromPrior(gen, col)
}

func (prob *planProblem) solveColumn(gen solutionGen, col *word.Column) {
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
		col.Fixed++
	}
	gen.Fix(c, v)
}

func (prob *planProblem) solveSingularColumn(gen solutionGen, col *word.Column) bool {
	if col.Have != 1 || col.Unknown != 1 {
		return false
	}

	if c := col.Chars[2]; col.I == 0 && c != 0 {
		// carry + _ + _ = c --> c == carry --> c = carry = 1
		if col.Prior == nil {
			panic("invalid final column: has no prior")
		}
		prob.fix(gen, c, 1)
		col.Solved = true
		col.Prior.Carry = word.CarryOne
		return true
	}

	return false
}

func (prob *planProblem) fixRange(min, max int, c byte) (int, int) {
	if min == 0 && (prob.fixedValues[0] ||
		c == prob.Words[0][0] ||
		c == prob.Words[1][0] ||
		c == prob.Words[2][0]) {
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

func (prob *planProblem) chooseOne(gen solutionGen, col *word.Column) bool {
	return prob.chooseFirst(gen, col)
}

func (prob *planProblem) chooseFirst(gen solutionGen, col *word.Column) bool {
	for _, cc := range col.Chars {
		if cc == 0 || prob.known[cc] {
			continue
		}
		min, max := prob.fixRange(0, prob.Base-1, cc)
		if min == max {
			prob.fix(gen, cc, min)
			return true
		}
		prob.chooseRange(gen, cc, min, max)
		return true
	}
	return false
}

func (prob *planProblem) chooseBest(gen solutionGen, col *word.Column) bool {
	var min, max, N [3]int

	i := -1
	for j, cc := range col.Chars {
		if cc == 0 || prob.known[cc] {
			continue
		}
		min[j], max[j] = prob.fixRange(0, prob.Base-1, cc)
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

	prob.chooseRange(gen, col.Chars[i], min[i], max[i])
	return true
}

func (prob *planProblem) chooseRange(gen solutionGen, c byte, min, max int) {
	gen.ChooseRange(c, min, max)
	prob.markKnown(c)
}

func (prob *planProblem) solveColumnFromPrior(gen solutionGen, col *word.Column) bool {
	if col.Prior != nil && col.Prior.Carry == word.CarryUnknown {
		// unknown prior carry is a case for assumeCarrySolveColumn
		return false
	}

	if col.Unknown == 0 {
		// this is checkColumn's job
		return false
	}

	if prob.annotated {
		gen.Logf("solveFromPrior: %s", col.Label())
	}

	for u := col.Unknown; u > 1; u = col.Unknown {
		if !prob.chooseOne(gen, col) {
			break
		}
	}
	if col.Unknown > 1 {
		// chooseOne was unable to figure it out
		return false
	}

	if c := col.Chars[0]; c != 0 && !prob.known[c] {
		gen.ComputeFirstSummand(col)
		col.Carry = word.CarryComputed
		prob.markKnown(c)
	} else if c := col.Chars[1]; c != 0 && !prob.known[c] {
		gen.ComputeSecondSummand(col)
		col.Carry = word.CarryComputed
		prob.markKnown(c)
	} else if c := col.Chars[2]; c != 0 && !prob.known[c] {
		gen.ComputeSum(col)
		col.Carry = word.CarryComputed
		prob.markKnown(c)
	} else {
		panic("invalid solveColumnFromPrior state")
	}

	col.Solved = true
	return true
}
