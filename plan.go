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
	planTopDown(prob, gen)
}

type solutionGen interface {
	init(plan planner, desc string)
	setCarry(plan planner, v int)
	fix(plan planner, c byte, v int)
	computeSum(plan planner, a, b, c byte)
	computeSummand(plan planner, a, b, c byte)
	computeCarry(plan planner, c1, c2 byte)
	choose(plan planner, c byte)
	checkColumn(plan planner, cx [3]byte)
	finish(plan planner)
}

type planner interface {
	problem() *planProblem
	knownLetters() map[byte]bool
}

type bottomUpPlan struct {
	prob *planProblem
	gen  solutionGen
}

type topDownPlan struct {
	bottomUpPlan
}

func planTopDown(prob *planProblem, gen solutionGen) {
	td := topDownPlan{bottomUpPlan{prob: prob, gen: gen}}
	td.gen.init(&td, "top down ... bottom up")
	td.plan()
}

func (td *topDownPlan) plan() {
	prob := td.bottomUpPlan.prob
	N := prob.numColumns()
	for i := 0; i < N; i++ {
		cx := prob.getColumn(i)

		if cx[0] == 0 && cx[1] == 0 && cx[2] != 0 && !prob.known[cx[2]] {
			td.gen.fix(td, cx[2], 1)
			prob.solved[i] = true
			prob.known[cx[2]] = true
			continue
		}
	}

	td.bottomUpPlan.plan()
}

func planBottomUp(prob *planProblem, gen solutionGen) {
	bu := bottomUpPlan{prob: prob, gen: gen}
	bu.gen.init(&bu, "bottom up")
	bu.plan()
}

func (bu *bottomUpPlan) plan() {
	prob := bu.prob
	// for each column from the right
	//   choose letters until 2/3 are known
	//   compute the third (if unknown)
	n := prob.numColumns() - 1
	var last [3]byte
	for i := n; i >= 0; i-- {
		cx := prob.getColumn(i)
		if i == n {
			bu.gen.setCarry(bu, 0)
		} else {
			bu.gen.computeCarry(bu, last[0], last[1])
		}
		bu.solveColumn(i, cx)
		last = cx
	}
	bu.gen.finish(bu)
}

func (bu *bottomUpPlan) problem() *planProblem {
	return bu.prob
}

func (bu *bottomUpPlan) knownLetters() map[byte]bool {
	return bu.prob.known
}

func (bu *bottomUpPlan) solveColumn(i int, cx [3]byte) {
	prob := bu.prob
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

	// TODO: hoist this call-site out to bu.plan once we reify a column struct
	// that can carry known counts, index, etc
	if numUnknown == 0 {
		bu.gen.checkColumn(bu, cx)
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
						bu.gen.computeSummand(bu, c, cx[1], cx[2])
					case 1:
						bu.gen.computeSummand(bu, c, cx[0], cx[2])
					case 2:
						bu.gen.computeSum(bu, cx[0], cx[1], c)
					}
				} else {
					bu.gen.choose(bu, c)
				}
				prob.known[c] = true
				numUnknown--
				numKnown++
			}
		}
	}

	prob.solved[i] = true
}
