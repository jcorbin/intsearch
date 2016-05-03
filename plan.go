package main

const (
	carryUnknown  = -1
	carryZero     = 0
	carryOne      = 1
	carryComputed = 2
)

type carryValue int

func (c carryValue) String() string {
	switch c {
	case carryUnknown:
		return "unknown"
	case carryZero:
		return "0"
	case carryOne:
		return "1"
	case carryComputed:
		return "computed"
	default:
		return "invalid"
	}
}

type column struct {
	i       int
	prior   *column
	cx      [3]byte
	solved  bool
	known   int
	unknown int
	carry   carryValue
}

type planProblem struct {
	problem
	columns []column
	letCols map[byte][]*column
	known   map[byte]bool
}

type solutionGen interface {
	init(desc string)
	fix(c byte, v int)
	fixCarry(i, v int)
	computeSum(col *column)
	computeFirstSummand(col *column)
	computeSecondSummand(col *column)
	choose(col *column, i int, c byte)
	checkColumn(col *column)
	finish()
}

func newPlanProblem(p *problem) *planProblem {
	C := p.numColumns()
	N := len(p.letterSet)
	prob := &planProblem{
		problem: *p,
		columns: make([]column, C),
		letCols: make(map[byte][]*column, N),
		known:   make(map[byte]bool, N),
	}
	var last *column
	for i := 0; i < C; i++ {
		col := &prob.columns[i]
		col.i = i
		col.carry = carryUnknown
		if last != nil {
			last.prior = col
		}
		col.cx = prob.getColumn(i)
		a, b, c := col.cx[0], col.cx[1], col.cx[2]
		if a != 0 {
			col.unknown++
			prob.letCols[a] = append(prob.letCols[a], col)
		}
		if b != 0 && b != a {
			col.unknown++
			prob.letCols[b] = append(prob.letCols[b], col)
		}
		if c != 0 && c != b && c != a {
			col.unknown++
			prob.letCols[c] = append(prob.letCols[c], col)
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
	prob.columns[0].carry = carryZero
	prob.planTopDown(gen)
}

func (prob *planProblem) planTopDown(gen solutionGen) {
	prob.procTopDown(gen, &prob.columns[0])
}

func (prob *planProblem) procTopDown(gen solutionGen, col *column) {
	a, b, c := col.cx[0], col.cx[1], col.cx[2]

	if col.i == 0 && a == 0 && b == 0 && c != 0 && !prob.known[c] {
		if col.prior == nil {
			panic("shouldn't be possible")
		}
		gen.fix(c, 1)
		col.solved = true
		prob.markKnown(c)
		prob.fixCarryIn(gen, col, carryOne)
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

func (prob *planProblem) solveColumn(gen solutionGen, col *column) {
	if col.unknown == 0 {
		gen.checkColumn(col)
		return
	}

	if col.solved {
		panic("invalid column solved state")
	}

	prob.solveColumnFromPrior(gen, col)

	if !col.solved {
		panic("cannot solve column")
	}
}

func (prob *planProblem) solveColumnFromPrior(gen solutionGen, col *column) {
	if col.prior != nil && col.prior.carry == carryUnknown {
		// unknown prior carry not yet support; i.e. solveColumn must be called
		// in bottom-up/right-to-left/decreasing-index order
		return
	}

	for x, c := range col.cx {
		if c != 0 && !prob.known[c] {
			if col.unknown == 1 {
				switch x {
				case 0:
					gen.computeFirstSummand(col)
				case 1:
					gen.computeSecondSummand(col)
				case 2:
					gen.computeSum(col)
				}
			} else {
				gen.choose(col, x, c)
			}
			prob.markKnown(c)
		}
	}
	col.solved = true
	col.carry = carryComputed
}
