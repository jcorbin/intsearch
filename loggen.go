package main

import (
	"fmt"
	"strings"
)

type logGen struct {
	*planProblem
	prefix   string
	step     int
	branches []int
}

func newLogGen(prob *planProblem) *logGen {
	return &logGen{
		planProblem: prob,
		prefix:      "// ",
		branches:    make([]int, 0, len(prob.letterSet)),
	}
}

func (lg *logGen) logf(format string, args ...interface{}) error {
	if _, err := fmt.Print(lg.prefix); err != nil {
		return err
	}
	if _, err := fmt.Printf(format, args...); err != nil {
		return err
	}
	return nil
}

func (lg *logGen) stepf(format string, args ...interface{}) {
	lg.step++
	format = fmt.Sprintf("step[%v]: %s", lg.step, format)
	lg.logf(format, args...)
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

func (lg *logGen) fix(c byte, v int) {
	lg.stepf("fix %v = %v\n", string(c), v)
}

func (lg *logGen) fixCarry(i, v int) {
	lg.stepf("fixCarry col[%d] = %v\n", i, v)
}

func (lg *logGen) computeSum(col *column) {
	a, b, c := col.cx[0], col.cx[1], col.cx[2]
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

func (lg *logGen) computeFirstSummand(col *column) {
	lg.computeSummand(col.cx[0], col.cx[1], col.cx[2])
}

func (lg *logGen) computeSecondSummand(col *column) {
	lg.computeSummand(col.cx[1], col.cx[0], col.cx[2])
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

func (lg *logGen) chooseRange(col *column, c byte, i, min, max int) {
	N := max - min
	R := lg.base - len(lg.known)
	if R < N {
		N = R
	}
	lg.branches = append(lg.branches, N)
	lg.stepf("choose %v (branch by %v)\n", string(c), N)
}

func (lg *logGen) checkColumn(col *column) {
	a, b, c := col.cx[0], col.cx[1], col.cx[2]
	if a != 0 && b != 0 {
		lg.stepf("check column carry + %v + %v = %v\n", string(a), string(b), string(c))
	} else if a != 0 {
		lg.stepf("check column carry + %v = %v\n", string(a), string(c))
	} else if b != 0 {
		lg.stepf("check column carry + %v = %v\n", string(b), string(c))
	} else {
		lg.stepf("check column carry = %v\n", string(c))
	}
}

func (lg *logGen) finish() {
	lg.stepf("done\n")

	branches := 1
	for _, b := range lg.branches {
		branches *= b
	}

	if lg.prefix == "" {
		fmt.Printf("// Total Branches: %v\n", branches)
	} else {
		fmt.Printf("// %s> Total Branches: %v\n", lg.prefix, branches)
	}
}
