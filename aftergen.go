package main

type afterGen struct {
	obs func(plan planner)
}

func (ag afterGen) init(plan planner, desc string) {
	ag.obs(plan)
}

func (ag afterGen) setCarry(plan planner, v int) {
	ag.obs(plan)
}

func (ag afterGen) fix(plan planner, c byte, v int) {
	ag.obs(plan)
}

func (ag afterGen) initColumn(plan planner, cx [3]byte, numKnown, numUnknown int) {
	ag.obs(plan)
}

func (ag afterGen) computeSum(plan planner, a, b, c byte) {
	ag.obs(plan)
}

func (ag afterGen) computeSummand(plan planner, a, b, c byte) {
	ag.obs(plan)
}

func (ag afterGen) computeCarry(plan planner, c1, c2 byte) {
	ag.obs(plan)
}

func (ag afterGen) choose(plan planner, c byte) {
	ag.obs(plan)
}

func (ag afterGen) checkFinal(plan planner, c byte, c1, c2 byte) {
	ag.obs(plan)
}

func (ag afterGen) finish(plan planner) {
	ag.obs(plan)
}
