package main

type multiGen []solutionGen

func (mg multiGen) logf(format string, args ...interface{}) error {
	for _, gen := range mg {
		gen.logf(format, args...)
	}
	return nil // TODO: multiError fwiw
}

func (mg multiGen) init(desc string) {
	for _, gen := range mg {
		gen.init(desc)
	}
}

func (mg multiGen) fix(c byte, v int) {
	for _, gen := range mg {
		gen.fix(c, v)
	}
}

func (mg multiGen) computeSum(col *column) {
	for _, gen := range mg {
		gen.computeSum(col)
	}
}

func (mg multiGen) computeFirstSummand(col *column) {
	for _, gen := range mg {
		gen.computeFirstSummand(col)
	}
}

func (mg multiGen) computeSecondSummand(col *column) {
	for _, gen := range mg {
		gen.computeSecondSummand(col)
	}
}

func (mg multiGen) chooseRange(col *column, c byte, i, min, max int) {
	for _, gen := range mg {
		gen.chooseRange(col, c, i, min, max)
	}
}

func (mg multiGen) checkColumn(col *column) {
	for _, gen := range mg {
		gen.checkColumn(col)
	}
}

func (mg multiGen) finish() {
	for _, gen := range mg {
		gen.finish()
	}
}

func (mg multiGen) finalize() {
	for _, gen := range mg {
		gen.finalize()
	}
}
