package main

import (
	"fmt"
	"strings"
)

type logGen struct {
	*planProblem
	step     int
	branches []int
}

func newLogGen(prob *planProblem) *logGen {
	return &logGen{
		planProblem: prob,
	}
}

func (lg *logGen) stepf(format string, args ...interface{}) {
	lg.step++
	fmt.Printf("// Step[%v]: ", lg.step)
	fmt.Printf(format, args...)
}

func (lg *logGen) init(desc string) {
	var w int
	for _, word := range lg.words {
		if len(word) > w {
			w = len(word)
		}
	}
	fmt.Printf("# Problem:\n")
	fmt.Printf("#   %s%v\n", strings.Repeat(" ", w-len(lg.words[0])), string(lg.words[0]))
	fmt.Printf("# + %s%v\n", strings.Repeat(" ", w-len(lg.words[1])), string(lg.words[1]))
	fmt.Printf("# = %s%v\n", strings.Repeat(" ", w-len(lg.words[2])), string(lg.words[2]))
	fmt.Printf("# base: %v\n", lg.base)
	fmt.Printf("# letters: %v\n", lg.sortedLetters())
	fmt.Printf("# method: %s\n", desc)
	fmt.Printf("\n")
}

func (lg *logGen) setCarry(v int) {
	lg.stepf("set carry = %d\n", v)
}

func (lg *logGen) fix(c byte, v int) {
	lg.stepf("fix %v = %v\n", string(c), v)
}

func (lg *logGen) computeSum(a, b, c byte) {
	if a != 0 && b != 0 {
		lg.stepf("compute %v = %v + %v + carry (mod %v)\n", string(c), string(a), string(b), lg.base)
	} else if a != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)\n", string(c), string(a), lg.base)
	} else if b != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)\n", string(c), string(b), lg.base)
	} else {
		lg.stepf("compute %v = carry (mod %v)\n", string(c), lg.base)
	}
}

func (lg *logGen) computeSummand(a, b, c byte) {
	if b != 0 && c != 0 {
		lg.stepf("compute %v = %v - %v - carry (mod %v)\n", string(a), string(b), string(c), lg.base)
	} else if b != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)\n", string(a), string(b), lg.base)
	} else if c != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)\n", string(a), string(c), lg.base)
	} else {
		lg.stepf("compute %v = - carry (mod %v)\n", string(a), lg.base)
	}
}

func (lg *logGen) computeCarry(c1, c2 byte) {
	if c1 != 0 && c2 != 0 {
		lg.stepf("set carry = (carry + %v + %v) // %v\n", string(c1), string(c2), lg.base)
	} else if c1 != 0 {
		lg.stepf("set carry = (carry + %v) // %v\n", string(c1), lg.base)
	} else if c2 != 0 {
		lg.stepf("set carry = (carry + %v) // %v\n", string(c2), lg.base)
	} else {
		lg.stepf("set carry = carry // %v = 0\n", lg.base)
	}
}

func (lg *logGen) choose(c byte) {
	branches := lg.base - len(lg.known)
	lg.branches = append(lg.branches, branches)
	lg.stepf("choose %v (branch by %v)\n", string(c), branches)
}

func (lg *logGen) checkColumn(cx [3]byte) {
	lg.stepf("check column %v + %v = %v\n", string(cx[0]), string(cx[1]), string(cx[2]))
}

func (lg *logGen) finish() {
	lg.stepf("done\n")

	branches := 1
	for _, b := range lg.branches {
		branches *= b
	}

	fmt.Printf("")
	fmt.Printf("# Total Branches: %v\n", branches)
}
