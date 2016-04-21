package main

type multiGen struct {
	gens []solutionGen
}

func (mg multiGen) init(plan planner, desc string) {
	for _, gen := range mg.gens {
		gen.init(plan, desc)
	}
}

func (mg multiGen) setCarry(plan planner, v int) {
	for _, gen := range mg.gens {
		gen.setCarry(plan, v)
	}
}

func (mg multiGen) fix(plan planner, c byte, v int) {
	for _, gen := range mg.gens {
		gen.fix(plan, c, v)
	}
}

func (mg multiGen) computeSum(plan planner, a, b, c byte) {
	for _, gen := range mg.gens {
		gen.computeSum(plan, a, b, c)
	}
}

func (mg multiGen) computeSummand(plan planner, a, b, c byte) {
	for _, gen := range mg.gens {
		gen.computeSummand(plan, a, b, c)
	}
}

func (mg multiGen) computeCarry(plan planner, c1, c2 byte) {
	for _, gen := range mg.gens {
		gen.computeCarry(plan, c1, c2)
	}
}

func (mg multiGen) choose(plan planner, c byte) {
	for _, gen := range mg.gens {
		gen.choose(plan, c)
	}
}

func (mg multiGen) checkColumn(plan planner, cx [3]byte) {
	for _, gen := range mg.gens {
		gen.checkColumn(plan, cx)
	}
}

func (mg multiGen) finish(plan planner) {
	for _, gen := range mg.gens {
		gen.finish(plan)
	}
}
