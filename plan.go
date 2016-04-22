package main

type planProblem struct {
	problem
	solved []bool
	known  map[byte]bool
}

func newPlanProblem(prob *problem) *planProblem {
	return &planProblem{
		problem: *prob,
		solved:  make([]bool, prob.numColumns()),
		known:   make(map[byte]bool, len(prob.letterSet)),
	}
}

func plan(prob *planProblem, gen solutionGen) {
	gen.init("top down ... bottom up")
	planTopDown(prob, gen)
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

func planTopDown(prob *planProblem, gen solutionGen) {
	N := prob.numColumns()
	for i := 0; i < N; i++ {
		cx := prob.getColumn(i)

		if cx[0] == 0 && cx[1] == 0 && cx[2] != 0 && !prob.known[cx[2]] {
			gen.fix(cx[2], 1)
			prob.solved[i] = true
			prob.known[cx[2]] = true
			continue
		}
	}

	planBottomUp(prob, gen)
}

func planBottomUp(prob *planProblem, gen solutionGen) {
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
		solveColumn(prob, gen, i, cx)
		last = cx
	}
	gen.finish()
}

func solveColumn(prob *planProblem, gen solutionGen, i int, cx [3]byte) {
	numKnown := 0
	numUnknown := 0
	for _, c := range cx {
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
		gen.checkColumn(cx)
		return
	}

	// TODO: reevaluate this check once we reify column struct
	if prob.solved[i] {
		// we have numUnknown > 0, but solved[i]
		panic("incorrect column solved note")
	}

	for x, c := range cx {
		if c != 0 {
			if !prob.known[c] {
				if numUnknown == 1 {
					switch x {
					case 0:
						gen.computeSummand(c, cx[1], cx[2])
					case 1:
						gen.computeSummand(c, cx[0], cx[2])
					case 2:
						gen.computeSum(cx[0], cx[1], c)
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

	prob.solved[i] = true
}
