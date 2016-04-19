package main

type solutionGen interface {
	init(plan planner, desc string)
	setCarry(plan planner, v int)
	fix(plan planner, c byte, v int)
	initColumn(plan planner, cx [3]byte, numKnown, numUnknown int)
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
	// for each column from the right
	//   choose letters until 2/3 are known
	//   compute the third (if unknown)
	p := bottomUpPlan{
		prob:  prob,
		gen:   gen,
		known: make(map[byte]bool, len(prob.letterSet)),
	}
	p.gen.init(&p, "bottom up")
	p.gen.setCarry(&p, 0)
	p.prob.eachColumn(p.solveColumn)
	p.gen.finish(&p)
}

func (p *bottomUpPlan) problem() *problem {
	return p.prob
}

func (p *bottomUpPlan) knownLetters() map[byte]bool {
	return p.known
}

func (p *bottomUpPlan) solveColumn(cx [3]byte) {
	numKnown := 0
	numUnknown := 0
	for _, c := range cx {
		if c != 0 {
			if p.known[c] {
				numKnown++
			}
			if !p.known[c] {
				numUnknown++
			}
		}
	}
	p.gen.initColumn(p, cx, numKnown, numUnknown)
	for x, c := range cx {
		if c != 0 {
			if !p.known[c] {
				if numUnknown == 1 {
					switch x {
					case 0:
						p.gen.computeSummand(p, c, cx[1], cx[2])
					case 1:
						p.gen.computeSummand(p, c, cx[0], cx[2])
					case 2:
						p.gen.computeSum(p, cx[0], cx[1], c)
					}
				} else {
					p.gen.choose(p, c)
				}
				p.known[c] = true
				numUnknown--
				numKnown++
			} else if x == 2 && cx[0] == 0 && cx[1] == 0 {
				p.gen.checkFinal(p, c, cx[0], cx[1])
			}
		}
	}
	p.gen.computeCarry(p, cx[0], cx[1])
}
