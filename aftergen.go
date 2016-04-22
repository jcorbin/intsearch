package main

type afterGen func(plan planner)

func (ag afterGen) init(plan planner, desc string) {
	ag(plan)
}

func (ag afterGen) setCarry(plan planner, v int) {
	ag(plan)
}

func (ag afterGen) fix(plan planner, c byte, v int) {
	ag(plan)
}

func (ag afterGen) computeSum(plan planner, a, b, c byte) {
	ag(plan)
}

func (ag afterGen) computeSummand(plan planner, a, b, c byte) {
	ag(plan)
}

func (ag afterGen) computeCarry(plan planner, c1, c2 byte) {
	ag(plan)
}

func (ag afterGen) choose(plan planner, c byte) {
	ag(plan)
}

func (ag afterGen) checkColumn(plan planner, cx [3]byte) {
	ag(plan)
}

func (ag afterGen) finish(plan planner) {
	ag(plan)
}
