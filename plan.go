package main

func plan(prob *problem, gen solutionGen) {
	planBottomUp(prob, gen)
}

type solutionGen interface {
	init(plan planner, desc string)
	setCarry(plan planner, v int)
	fix(plan planner, c byte, v int)
	computeSum(plan planner, a, b, c byte)
	computeSummand(plan planner, a, b, c byte)
	computeCarry(plan planner, c1, c2 byte)
	choose(plan planner, c byte)
	checkFinal(plan planner, c byte, c1, c2 byte)
	finish(plan planner)
}

type planner interface {
	problem() *problem
	knownLetters() map[byte]bool
}

type bottomUpPlan struct {
	prob  *problem
	gen   solutionGen
	known map[byte]bool
}

func planBottomUp(prob *problem, gen solutionGen) {
	bu := bottomUpPlan{
		prob:  prob,
		gen:   gen,
		known: make(map[byte]bool, len(prob.letterSet)),
	}
	bu.gen.init(&bu, "bottom up")
	bu.plan()
}

func (bu *bottomUpPlan) plan() {
	// for each column from the right
	//   choose letters until 2/3 are known
	//   compute the third (if unknown)
	n := bu.prob.numColumns() - 1
	var last [3]byte
	for i := n; i >= 0; i-- {
		cx := bu.prob.getColumn(i)
		if i == n {
			bu.gen.setCarry(bu, 0)
		} else {
			bu.gen.computeCarry(bu, last[0], last[1])
		}
		bu.solveColumn(cx)
		last = cx
	}
	bu.gen.finish(bu)
}

func (bu *bottomUpPlan) problem() *problem {
	return bu.prob
}

func (bu *bottomUpPlan) knownLetters() map[byte]bool {
	return bu.known
}

func (bu *bottomUpPlan) solveColumn(cx [3]byte) {
	numKnown := 0
	numUnknown := 0
	for _, c := range cx {
		if c != 0 {
			if bu.known[c] {
				numKnown++
			}
			if !bu.known[c] {
				numUnknown++
			}
		}
	}
	for x, c := range cx {
		if c != 0 {
			if !bu.known[c] {
				if numUnknown == 1 {
					switch x {
					case 0:
						bu.gen.computeSummand(bu, c, cx[1], cx[2])
					case 1:
						bu.gen.computeSummand(bu, c, cx[0], cx[2])
					case 2:
						bu.gen.computeSum(bu, cx[0], cx[1], c)
					}
				} else {
					bu.gen.choose(bu, c)
				}
				bu.known[c] = true
				numUnknown--
				numKnown++
			} else if x == 2 && cx[0] == 0 && cx[1] == 0 {
				bu.gen.checkFinal(bu, c, cx[0], cx[1])
			}
		}
	}
}
