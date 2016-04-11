package main

type afterGen struct {
	obs func(prob *problem)
}

func (ag afterGen) init(prob *problem, desc string) {
	ag.obs(prob)
}

func (ag afterGen) fix(prob *problem, c byte, v int) {
	ag.obs(prob)
}

func (ag afterGen) interColumn(prob *problem, cx [3]byte) {
	ag.obs(prob)
}

func (ag afterGen) initColumn(prob *problem, cx [3]byte, numKnown, numUnknown int) {
	ag.obs(prob)
}

func (ag afterGen) solve(prob *problem, neg bool, c byte, c1, c2 byte) {
	ag.obs(prob)
}

func (ag afterGen) choose(prob *problem, c byte) {
	ag.obs(prob)
}

func (ag afterGen) checkFinal(prob *problem, c byte, c1, c2 byte) {
	ag.obs(prob)
}

func (ag afterGen) finish(prob *problem) {
	ag.obs(prob)
}
