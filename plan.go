package main

import "log"

type carryValue int

const (
	carryUnknown carryValue = iota - 1
	carryZero
	carryOne
	carryComputed
)

//go:generate stringer -type=carryValue

type column struct {
	i       int
	prior   *column
	cx      [3]byte
	solved  bool
	have    int
	known   int
	unknown int
	carry   carryValue
}

func (col *column) priorCarry() carryValue {
	if col.prior == nil {
		return carryZero
	}
	return col.prior.carry
}

type planProblem struct {
	problem
	columns     []column
	letCols     map[byte][]*column
	known       map[byte]bool
	fixedValues []bool
}

type solutionGen interface {
	logf(string, ...interface{}) error
	init(desc string)
	fix(c byte, v int)
	fixCarry(i, v int)
	computeSum(col *column)
	computeFirstSummand(col *column)
	computeSecondSummand(col *column)
	chooseRange(col *column, c byte, i, min, max int)
	checkColumn(col *column)
	finish()
}

func newPlanProblem(p *problem) *planProblem {
	C := p.numColumns()
	N := len(p.letterSet)
	prob := &planProblem{
		problem:     *p,
		columns:     make([]column, C),
		letCols:     make(map[byte][]*column, N),
		known:       make(map[byte]bool, N),
		fixedValues: make([]bool, p.base),
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

func (prob *planProblem) plan(gen solutionGen) {
	gen.init("top down ... bottom up")
	prob.planTopDown(gen)
}

func (prob *planProblem) planTopDown(gen solutionGen) {
	prob.procTopDown(gen, &prob.columns[0])
}

func (prob *planProblem) procTopDown(gen solutionGen, col *column) {
	if col.prior == nil {
		prob.solveColumn(gen, col)
		gen.finish()
		return
	}

	if prob.maySolveColumn(gen, col) {
		prob.procTopDown(gen, col.prior)
		return
	}

	prob.planBottomUp(gen)
}

func (prob *planProblem) fixCarryIn(gen solutionGen, col *column, carry carryValue) {
	switch carry {
	case carryZero:
		fallthrough
	case carryOne:
		col.prior.carry = carry
		gen.fixCarry(col.i, int(carry))
	default:
		panic("can only fix carry to a definite value")
	}
}

func (prob *planProblem) planBottomUp(gen solutionGen) {
	for i := prob.numColumns() - 1; i >= 0; i-- {
		prob.solveColumn(gen, &prob.columns[i])
	}
	gen.finish()
}

func (prob *planProblem) checkColumn(gen solutionGen, col *column) bool {
	if !col.solved {
		gen.checkColumn(col)
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
	prob.fixedValues[v] = true
	gen.fix(c, v)
	prob.markKnown(c)
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
		prob.fixCarryIn(gen, col, carryOne)
		return true
	}

	return false
}

func (prob *planProblem) choiceRange(col *column, c byte, i int) (int, int) {
	min, max := 0, prob.base-1

	if prob.fixedValues[0] ||
		c == prob.words[0][0] ||
		c == prob.words[1][0] ||
		c == prob.words[2][0] {
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
	var (
		c              byte
		i, min, max, N int
	)

	for j, cc := range col.cx {
		if cc == 0 || prob.known[cc] {
			continue
		}

		cMin, cMax := prob.choiceRange(col, cc, j)
		if cMin == cMax {
			prob.fix(gen, cc, cMin)
			return true
		}

		if c == 0 {
			c, i, min, max, N = cc, j, cMin, cMax, cMax-cMin
			continue
		}

		if M := cMax - cMin; M < N {
			c, i, min, max, N = cc, j, cMin, cMax, M
		}
	}

	if c == 0 {
		return false
	}

	gen.chooseRange(col, c, i, min, max)
	prob.markKnown(c)
	return true
}

func (prob *planProblem) solveColumnFromPrior(gen solutionGen, col *column) bool {
	if col.priorCarry() == carryUnknown {
		// unknown prior carry not yet support; i.e. solveColumn must be called
		// in bottom-up/right-to-left/decreasing-index order
		return false
	} else if col.unknown == 0 {
		// this is checkColumn's job
		return false
	}

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
