package main

type afterGen func()

func (ag afterGen) init(desc string) {
	ag()
}

func (ag afterGen) setCarry(v int) {
	ag()
}

func (ag afterGen) fix(c byte, v int) {
	ag()
}

func (ag afterGen) computeSum(a, b, c byte) {
	ag()
}

func (ag afterGen) computeSummand(a, b, c byte) {
	ag()
}

func (ag afterGen) computeCarry(c1, c2 byte) {
	ag()
}

func (ag afterGen) choose(c byte) {
	ag()
}

func (ag afterGen) checkColumn(cx [3]byte) {
	ag()
}

func (ag afterGen) finish() {
	ag()
}
