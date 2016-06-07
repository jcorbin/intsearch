package main

import (
	"fmt"
	"strings"

	"github.com/jcorbin/intsearch/word"
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
		branches:    make([]int, 0, len(prob.Letters)),
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
	for _, word := range lg.Words {
		if len(word) > w {
			w = len(word)
		}
	}
	letters := make([]string, len(lg.Letters))
	for i, c := range lg.SortedLetters() {
		letters[i] = string(c)
	}
	lg.logf("Problem:")
	lg.logf("  %s%v", strings.Repeat(" ", w-len(lg.Words[0])), string(lg.Words[0]))
	lg.logf("+ %s%v", strings.Repeat(" ", w-len(lg.Words[1])), string(lg.Words[1]))
	lg.logf("= %s%v", strings.Repeat(" ", w-len(lg.Words[2])), string(lg.Words[2]))
	lg.logf("base: %v", lg.Base)
	lg.logf("letters: %v", letters)
	lg.logf("method: %s", desc)
	lg.logf("")
}

func (lg *logGen) fork(prob *planProblem, name, alt, cont string) solutionGen {
	if alt == "" {
		alt = fmt.Sprintf("%s:alt", name)
	}
	if cont == "" {
		cont = fmt.Sprintf("%s:cont", name)
	}
	n := 0
	for n < len(lg.prefix)-1 && lg.prefix[n] == ' ' {
		n++
	}
	lg.prefix = fmt.Sprintf("%s%s", strings.Repeat(" ", n), cont)
	return &logGen{
		planProblem: prob,
		prefix:      fmt.Sprintf("%s%s", strings.Repeat(" ", n+2), alt),
		step:        lg.step,
		branches:    lg.branches,
	}
}

func (lg *logGen) fix(c byte, v int) {
	lg.stepf("fix %v = %v", string(c), v)
}

func (lg *logGen) computeSum(col *word.Column) {
	a, b, c := col.Chars[0], col.Chars[1], col.Chars[2]
	if a != 0 && b != 0 {
		lg.stepf("compute %v = %v + %v + carry (mod %v)", string(c), string(a), string(b), lg.Base)
	} else if a != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)", string(c), string(a), lg.Base)
	} else if b != 0 {
		lg.stepf("compute %v = %v + carry (mod %v)", string(c), string(b), lg.Base)
	} else {
		lg.stepf("compute %v = carry (mod %v)", string(c), lg.Base)
	}
}

func (lg *logGen) computeFirstSummand(col *word.Column) {
	lg.computeSummand(col.Chars[0], col.Chars[1], col.Chars[2])
}

func (lg *logGen) computeSecondSummand(col *word.Column) {
	lg.computeSummand(col.Chars[1], col.Chars[0], col.Chars[2])
}

func (lg *logGen) computeSummand(a, b, c byte) {
	if b != 0 && c != 0 {
		lg.stepf("compute %v = %v - %v - carry (mod %v)", string(a), string(b), string(c), lg.Base)
	} else if b != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)", string(a), string(b), lg.Base)
	} else if c != 0 {
		lg.stepf("compute %v = %v - carry (mod %v)", string(a), string(c), lg.Base)
	} else {
		lg.stepf("compute %v = - carry (mod %v)", string(a), lg.Base)
	}
}

func (lg *logGen) chooseRange(c byte, min, max int) {
	N := max - min
	R := lg.Base - len(lg.known)
	if R < N {
		N = R
	}
	lg.branches = append(lg.branches, N)
	lg.stepf("choose %v (branch by %v)", string(c), N)
}

func (lg *logGen) checkColumn(col *word.Column, err error) {
	lg.stepf("check column: %s", col.Label())
}

func (lg *logGen) verify() {
	lg.stepf("verify")
}

func (lg *logGen) check(err error) {
	lg.stepf("check")
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
