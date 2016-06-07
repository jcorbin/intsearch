package main

import "github.com/jcorbin/intsearch/word"

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

func (mg multiGen) fork(prob *planProblem, name, alt, cont string) solutionGen {
	altGen := make([]solutionGen, len(mg))
	for i, gen := range mg {
		altGen[i] = gen.fork(prob, name, alt, cont)
	}
	return multiGen(altGen)
}

func (mg multiGen) fix(c byte, v int) {
	for _, gen := range mg {
		gen.fix(c, v)
	}
}

func (mg multiGen) computeSum(col *word.Column) {
	for _, gen := range mg {
		gen.computeSum(col)
	}
}

func (mg multiGen) computeFirstSummand(col *word.Column) {
	for _, gen := range mg {
		gen.computeFirstSummand(col)
	}
}

func (mg multiGen) computeSecondSummand(col *word.Column) {
	for _, gen := range mg {
		gen.computeSecondSummand(col)
	}
}

func (mg multiGen) chooseRange(c byte, min, max int) {
	for _, gen := range mg {
		gen.chooseRange(c, min, max)
	}
}

func (mg multiGen) checkColumn(col *word.Column, err error) {
	for _, gen := range mg {
		gen.checkColumn(col, err)
	}
}

func (mg multiGen) verify() {
	for _, gen := range mg {
		gen.verify()
	}
}

func (mg multiGen) check(err error) {
	for _, gen := range mg {
		gen.check(err)
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
