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

func (lg *logGen) init(prob *problem, desc string) {
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
	lg.stepf("set carry = 0\n")
}

func (lg *logGen) fix(prob *problem, c rune, v int) {
	lg.stepf("fix %v = %v\n", string(c), v)
}

func (lg *logGen) interColumn(prob *problem, cx [3]rune) {
	if cx[0] != 0 && cx[1] != 0 {
		lg.stepf("set carry = (%v + %v + carry) // %v\n", string(cx[0]), string(cx[1]), prob.base)
	} else if cx[0] != 0 {
		lg.stepf("set carry = (%v + carry) // %v\n", string(cx[0]), prob.base)
	} else if cx[1] != 0 {
		lg.stepf("set carry = (%v + carry) // %v\n", string(cx[1]), prob.base)
	} else {
		lg.stepf("set carry = carry // %v\n", prob.base)
	}
}

func (lg *logGen) initColumn(prob *problem, cx [3]rune, numKnown, numUnknown int) {
	if cx[0] != 0 && cx[1] != 0 {
		fmt.Printf("// column: carry + %v + %v = %v\n", string(cx[0]), string(cx[1]), string(cx[2]))
	} else if cx[0] != 0 {
		fmt.Printf("// column: carry + %v = %v\n", string(cx[0]), string(cx[2]))
	} else if cx[1] != 0 {
		fmt.Printf("// column: carry + %v = %v\n", string(cx[1]), string(cx[2]))
	}
}

func (lg *logGen) solve(prob *problem, neg bool, c, c1, c2 rune) {
	if c1 != 0 && c2 != 0 {
		if neg {
			lg.stepf("solve %v = %v - %v - carry (mod %v)\n", string(c), string(c1), string(c2), prob.base)
		} else {
			lg.stepf("solve %v = %v + %v + carry (mod %v)\n", string(c), string(c1), string(c2), prob.base)
		}
	} else if c1 != 0 {
		if neg {
			lg.stepf("solve %v = %v - carry (mod %v)\n", string(c), string(c1), prob.base)
		} else {
			lg.stepf("solve %v = %v + carry (mod %v)\n", string(c), string(c1), prob.base)
		}
	} else if c2 != 0 {
		if neg {
			lg.stepf("solve %v = %v - carry (mod %v)\n", string(c), string(c2), prob.base)
		} else {
			lg.stepf("solve %v = %v + carry (mod %v)\n", string(c), string(c2), prob.base)
		}
	} else {
		if neg {
			lg.stepf("solve %v = - carry (mod %v)\n", string(c), prob.base)
		} else {
			lg.stepf("solve %v = + carry (mod %v)\n", string(c), prob.base)
		}
	}
}

func (lg *logGen) choose(prob *problem, c rune) {
	branches := prob.base - len(prob.known)
	lg.branches = append(lg.branches, branches)
	lg.stepf("choose %v (branch by %v)\n", string(c), branches)
}

func (lg *logGen) checkFinal(prob *problem, c, c1, c2 rune) {
	lg.stepf("check %v == carry\n", string(c))
}

func (lg *logGen) finish(prob *problem) {
	lg.stepf("done\n")

	branches := 1
	for _, b := range lg.branches {
		branches *= b
	}

	fmt.Printf("")
	fmt.Printf("# Total Branches: %v\n", branches)
}
