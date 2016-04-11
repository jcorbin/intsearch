package main

type multiGen struct {
	gens []solutionGen
}

func (mg multiGen) init(prob *problem, desc string) {
	for _, gen := range mg.gens {
		gen.init(prob, desc)
	}
}

func (mg multiGen) fix(prob *problem, c byte, v int) {
	for _, gen := range mg.gens {
		gen.fix(prob, c, v)
	}
}

func (mg multiGen) interColumn(prob *problem, cx [3]byte) {
	for _, gen := range mg.gens {
		gen.interColumn(prob, cx)
	}
}

func (mg multiGen) initColumn(prob *problem, cx [3]byte, numKnown, numUnknown int) {
	for _, gen := range mg.gens {
		gen.initColumn(prob, cx, numKnown, numUnknown)
	}
}

func (mg multiGen) solve(prob *problem, neg bool, c byte, c1, c2 byte) {
	for _, gen := range mg.gens {
		gen.solve(prob, neg, c, c1, c2)
	}
}

func (mg multiGen) choose(prob *problem, c byte) {
	for _, gen := range mg.gens {
		gen.choose(prob, c)
	}
}

func (mg multiGen) checkFinal(prob *problem, c byte, c1, c2 byte) {
	for _, gen := range mg.gens {
		gen.checkFinal(prob, c, c1, c2)
	}
}

func (mg multiGen) finish(prob *problem) {
	for _, gen := range mg.gens {
		gen.finish(prob)
	}
}
