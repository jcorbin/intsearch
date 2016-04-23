package main

type afterGen func()

func (ag afterGen) init(desc string) {
	ag()
}

func (ag afterGen) fix(c byte, v int) {
	ag()
}

func (ag afterGen) computeSum(col *column) {
	ag()
}

func (ag afterGen) computeFirstSummand(col *column) {
	ag()
}

func (ag afterGen) computeSecondSummand(col *column) {
	ag()
}

func (ag afterGen) choose(c byte) {
	ag()
}

func (ag afterGen) checkColumn(col *column) {
	ag()
}

func (ag afterGen) finish() {
	ag()
}
