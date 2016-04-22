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

func (mg multiGen) computeSum(a, b, c byte) {
	for _, gen := range mg {
		gen.computeSum(a, b, c)
	}
}

func (mg multiGen) computeSummand(a, b, c byte) {
	for _, gen := range mg {
		gen.computeSummand(a, b, c)
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

func (mg multiGen) checkColumn(cx [3]byte) {
	for _, gen := range mg {
		gen.checkColumn(cx)
	}
}

func (mg multiGen) finish() {
	for _, gen := range mg {
		gen.finish()
	}
}
