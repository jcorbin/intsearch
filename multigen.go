package main

type multiGen []solutionGen

func (mg multiGen) init(desc string) {
	for _, gen := range mg {
		gen.init(desc)
	}
}

func (mg multiGen) setCarry(v int) {
	for _, gen := range mg {
		gen.setCarry(v)
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

func (mg multiGen) computeCarry(c1, c2 byte) {
	for _, gen := range mg {
		gen.computeCarry(c1, c2)
	}
}

func (mg multiGen) choose(c byte) {
	for _, gen := range mg {
		gen.choose(c)
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
