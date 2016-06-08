package main

import "github.com/jcorbin/intsearch/word"

type multiGen []word.SolutionGen

func (mg multiGen) Logf(format string, args ...interface{}) error {
	for _, gen := range mg {
		gen.Logf(format, args...)
	}
	return nil // TODO: multiError fwiw
}

func (mg multiGen) Init(desc string) {
	for _, gen := range mg {
		gen.Init(desc)
	}
}

func (mg multiGen) Fork(prob *word.PlanProblem, name, alt, cont string) word.SolutionGen {
	altGen := make([]word.SolutionGen, len(mg))
	for i, gen := range mg {
		altGen[i] = gen.Fork(prob, name, alt, cont)
	}
	return multiGen(altGen)
}

func (mg multiGen) Fix(c byte, v int) {
	for _, gen := range mg {
		gen.Fix(c, v)
	}
}

func (mg multiGen) ComputeSum(col *word.Column) {
	for _, gen := range mg {
		gen.ComputeSum(col)
	}
}

func (mg multiGen) ComputeFirstSummand(col *word.Column) {
	for _, gen := range mg {
		gen.ComputeFirstSummand(col)
	}
}

func (mg multiGen) ComputeSecondSummand(col *word.Column) {
	for _, gen := range mg {
		gen.ComputeSecondSummand(col)
	}
}

func (mg multiGen) ChooseRange(c byte, min, max int) {
	for _, gen := range mg {
		gen.ChooseRange(c, min, max)
	}
}

func (mg multiGen) CheckColumn(col *word.Column, err error) {
	for _, gen := range mg {
		gen.CheckColumn(col, err)
	}
}

func (mg multiGen) Verify() {
	for _, gen := range mg {
		gen.Verify()
	}
}

func (mg multiGen) Check(err error) {
	for _, gen := range mg {
		gen.Check(err)
	}
}

func (mg multiGen) Finish() {
	for _, gen := range mg {
		gen.Finish()
	}
}

func (mg multiGen) Finalize() {
	for _, gen := range mg {
		gen.Finalize()
	}
}
