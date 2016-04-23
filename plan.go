package main

type column struct {
	i      int
	cx     [3]byte
	solved bool
}

type planProblem struct {
	problem
	columns []column
	known   map[byte]bool
}

type solutionGen interface {
	init(desc string)
	setCarry(v int)
	fix(c byte, v int)
	computeSum(a, b, c byte)
	computeSummand(a, b, c byte)
	computeCarry(c1, c2 byte)
	choose(c byte)
	checkColumn(cx [3]byte)
	finish()
}

func newPlanProblem(p *problem) *planProblem {
	C := p.numColumns()
	N := len(p.letterSet)
	prob := &planProblem{
		problem: *p,
		columns: make([]column, C),
		known:   make(map[byte]bool, N),
	}
	for i := 0; i < C; i++ {
		col := &prob.columns[i]
		col.i = i
		col.cx = prob.getColumn(i)
	}
	return prob
}

func (prob *planProblem) plan(gen solutionGen) {
	gen.init("top down ... bottom up")
	prob.planTopDown(gen)
}

func (prob *planProblem) planTopDown(gen solutionGen) {
	N := prob.numColumns()
	for i := 0; i < N; i++ {
		col := &prob.columns[i]
		a, b, c := col.cx[0], col.cx[1], col.cx[2]

		if a == 0 && b == 0 && c != 0 && !prob.known[c] {
			gen.fix(c, 1)
			col.solved = true
			prob.known[c] = true
			continue
		}
	}

	prob.planBottomUp(gen)
}

func (prob *planProblem) planBottomUp(gen solutionGen) {
	// for each column from the right
	//   choose letters until 2/3 are known
	//   compute the third (if unknown)
	n := prob.numColumns() - 1
	var last [3]byte
	for i := n; i >= 0; i-- {
		cx := prob.getColumn(i)
		if i == n {
			gen.setCarry(0)
		} else {
			gen.computeCarry(last[0], last[1])
		}
		col := &prob.columns[i]
		prob.solveColumn(gen, col)
		last = cx
	}
	gen.finish()
}

func (prob *planProblem) solveColumn(gen solutionGen, col *column) {
	numKnown := 0
	numUnknown := 0
	for _, c := range col.cx {
		if c != 0 {
			if prob.known[c] {
				numKnown++
			}
			if !prob.known[c] {
				numUnknown++
			}
		}
	}

	// TODO: hoist this call-site out to planBottomUp once we reify a column
	// struct that can carry known counts, index, etc
	if numUnknown == 0 {
		gen.checkColumn(col.cx)
		return
	}

	// TODO: reevaluate this check once we reify column struct
	if col.solved {
		// we have numUnknown > 0, but solved
		panic("incorrect column solved note")
	}

	for x, c := range col.cx {
		if c != 0 {
			if !prob.known[c] {
				if numUnknown == 1 {
					switch x {
					case 0:
						gen.computeSummand(c, col.cx[1], col.cx[2])
					case 1:
						gen.computeSummand(c, col.cx[0], col.cx[2])
					case 2:
						gen.computeSum(col.cx[0], col.cx[1], c)
					}
				} else {
					gen.choose(c)
				}
				prob.known[c] = true
				numUnknown--
				numKnown++
			}
		}
	}

	col.solved = true
}
