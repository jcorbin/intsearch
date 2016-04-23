package main

type column struct {
	i       int
	prior   *column
	cx      [3]byte
	solved  bool
	known   int
	unknown int
}

type planProblem struct {
	problem
	columns []column
	letCols map[byte][]*column
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
		letCols: make(map[byte][]*column, N),
		known:   make(map[byte]bool, N),
	}
	var last *column
	for i := 0; i < C; i++ {
		col := &prob.columns[i]
		col.i = i
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
			prob.markKnown(c)
			continue
		}
	}

	prob.planBottomUp(gen)
}

func (prob *planProblem) planBottomUp(gen solutionGen) {
	for i := prob.numColumns() - 1; i >= 0; i-- {
		prob.solveColumn(gen, &prob.columns[i])
	}
	gen.finish()
}

func (prob *planProblem) solveColumn(gen solutionGen, col *column) {
	if pri := col.prior; pri == nil {
		gen.setCarry(0)
	} else {
		gen.computeCarry(pri.cx[0], pri.cx[1])
	}

	if col.unknown == 0 {
		gen.checkColumn(col.cx)
		return
	}

	if col.solved {
		panic("invalid column solved state")
	}

	for x, c := range col.cx {
		if c != 0 && !prob.known[c] {
			if col.unknown == 1 {
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
			prob.markKnown(c)
		}
	}

	col.solved = true
}
