package main

type afterGen func()

func (ag afterGen) logf(format string, args ...interface{}) error {
	return nil
}

func (ag afterGen) init(desc string) {
	ag()
}

func (ag afterGen) fork(prob *planProblem, name, alt, cont string) solutionGen {
	ag()
	return ag
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

func (ag afterGen) chooseRange(c byte, min, max int) {
	ag()
}

func (ag afterGen) checkColumn(col *column, err error) {
	ag()
}

func (ag afterGen) verify() {
	ag()
}

func (ag afterGen) check(err error) {
	ag()
}

func (ag afterGen) finish() {
	ag()
}

func (ag afterGen) finalize() {
	ag()
}
