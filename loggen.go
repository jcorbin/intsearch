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
		prefix:      "",
		branches:    make([]int, 0, len(prob.letterSet)),
	}
}

func (lg *logGen) logf(format string, args ...interface{}) error {
	if len(lg.prefix) == 0 {
		format = fmt.Sprintf("// %s\n", format)
	} else {
		format = fmt.Sprintf("// %s> %s\n", lg.prefix, format)
	}
	_, err := fmt.Printf(format, args...)
	return err
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
	letters := make([]string, len(lg.letterSet))
	for i, c := range lg.sortedLetters() {
		letters[i] = string(c)
	}
	lg.logf("Problem:")
	lg.logf("  %s%v", strings.Repeat(" ", w-len(lg.words[0])), string(lg.words[0]))
	lg.logf("+ %s%v", strings.Repeat(" ", w-len(lg.words[1])), string(lg.words[1]))
	lg.logf("= %s%v", strings.Repeat(" ", w-len(lg.words[2])), string(lg.words[2]))
	lg.logf("base: %v", lg.base)
	lg.logf("letters: %v", letters)
	lg.logf("method: %s", desc)
	lg.logf("")
}

func (lg *logGen) fix(c byte, v int) {
	lg.stepf("fix %v = %v", string(c), v)
}

func (lg *logGen) computeSum(col *column) {
	a, b, c := col.cx[0], col.cx[1], col.cx[2]
	if a != 0 && b != 0 {
		lg.stepf("compute %v = %v + %v + carry (mod %v)", string(c), string(a), string(b), lg.base)
	} else if a != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)", string(c), string(a), lg.base)
	} else if b != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)", string(c), string(b), lg.base)
	} else {
		lg.stepf("compute %v = carry (mod %v)", string(c), lg.base)
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
		lg.stepf("compute %v = %v - %v - carry (mod %v)", string(a), string(b), string(c), lg.base)
	} else if b != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)", string(a), string(b), lg.base)
	} else if c != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)", string(a), string(c), lg.base)
	} else {
		lg.stepf("compute %v = - carry (mod %v)", string(a), lg.base)
	}
}

func (lg *logGen) chooseRange(col *column, c byte, i, min, max int) {
	N := max - min
	R := lg.base - len(lg.known)
	if R < N {
		N = R
	}
	lg.branches = append(lg.branches, N)
	lg.stepf("choose %v (branch by %v)", string(c), N)
}

func (lg *logGen) checkColumn(col *column) {
	lg.stepf("check column: %s", col.label())
}

func (lg *logGen) verify() {
	lg.stepf("verify")
}

func (lg *logGen) finish() {
	lg.stepf("finish")
}

func (lg *logGen) finalize() {
	branches := 1
	for _, b := range lg.branches {
		branches *= b
	}

	lg.logf("Total Branches: %v", branches)
}
