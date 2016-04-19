package main

import (
	"fmt"
	"strings"
)

type logGen struct {
	step     int
	branches []int
}

func (lg *logGen) stepf(format string, args ...interface{}) {
	lg.step++
	fmt.Printf("// Step[%v]: ", lg.step)
	fmt.Printf(format, args...)
}

func (lg *logGen) init(plan planner, desc string) {
	prob := plan.problem()
	var w int
	for _, word := range prob.words {
		if len(word) > w {
			w = len(word)
		}
	}
	fmt.Printf("# Problem:\n")
	fmt.Printf("#   %s%v\n", strings.Repeat(" ", w-len(prob.words[0])), string(prob.words[0]))
	fmt.Printf("# + %s%v\n", strings.Repeat(" ", w-len(prob.words[1])), string(prob.words[1]))
	fmt.Printf("# = %s%v\n", strings.Repeat(" ", w-len(prob.words[2])), string(prob.words[2]))
	fmt.Printf("# base: %v\n", prob.base)
	fmt.Printf("# letters: %v\n", prob.sortedLetters())
	fmt.Printf("# method: %s\n", desc)
	fmt.Printf("\n")
}

func (lg *logGen) setCarry(plan planner, v int) {
	lg.stepf("set carry = %d\n", v)
}

func (lg *logGen) fix(plan planner, c byte, v int) {
	lg.stepf("fix %v = %v\n", string(c), v)
}

func (lg *logGen) computeSum(plan planner, a, b, c byte) {
	prob := plan.problem()
	if a != 0 && b != 0 {
		lg.stepf("compute %v = %v + %v + carry (mod %v)\n", string(c), string(a), string(b), prob.base)
	} else if a != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)\n", string(c), string(a), prob.base)
	} else if b != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)\n", string(c), string(b), prob.base)
	} else {
		lg.stepf("compute %v = carry (mod %v)\n", string(c), prob.base)
	}
}

func (lg *logGen) computeSummand(plan planner, a, b, c byte) {
	prob := plan.problem()
	if b != 0 && c != 0 {
		lg.stepf("compute %v = %v - %v - carry (mod %v)\n", string(a), string(b), string(c), prob.base)
	} else if b != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)\n", string(a), string(b), prob.base)
	} else if c != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)\n", string(a), string(c), prob.base)
	} else {
		lg.stepf("compute %v = - carry (mod %v)\n", string(a), prob.base)
	}
}

func (lg *logGen) computeCarry(plan planner, c1, c2 byte) {
	prob := plan.problem()
	if c1 != 0 && c2 != 0 {
		lg.stepf("set carry = (carry + %v + %v) // %v\n", string(c1), string(c2), prob.base)
	} else if c1 != 0 {
		lg.stepf("set carry = (carry + %v) // %v\n", string(c1), prob.base)
	} else if c2 != 0 {
		lg.stepf("set carry = (carry + %v) // %v\n", string(c2), prob.base)
	} else {
		lg.stepf("set carry = carry // %v = 0\n", prob.base)
	}
}

func (lg *logGen) choose(plan planner, c byte) {
	prob := plan.problem()
	branches := prob.base - len(plan.knownLetters())
	lg.branches = append(lg.branches, branches)
	lg.stepf("choose %v (branch by %v)\n", string(c), branches)
}

func (lg *logGen) checkFinal(plan planner, c, c1, c2 byte) {
	lg.stepf("check %v == carry\n", string(c))
}

func (lg *logGen) finish(plan planner) {
	lg.stepf("done\n")

	branches := 1
	for _, b := range lg.branches {
		branches *= b
	}

	fmt.Printf("")
	fmt.Printf("# Total Branches: %v\n", branches)
}
