package word

import (
	"fmt"
	"strings"
)

// LogGen implements a log observability SolutionGen that prints informative
// messages.  A prefix is provided that grows with fork context.
type LogGen struct {
	*PlanProblem
	prefix   string
	step     int
	branches []int
}

// NewLogGen creates a new LogGen for a given problem being planned.
func NewLogGen(prob *PlanProblem) *LogGen {
	return &LogGen{
		PlanProblem: prob,
		prefix:      "",
		branches:    make([]int, 0, len(prob.Letters)),
	}
}

// Logf simply formats and prints the passed message with an added prefix.
func (lg *LogGen) Logf(format string, args ...interface{}) error {
	if len(lg.prefix) == 0 {
		format = fmt.Sprintf("// %s\n", format)
	} else {
		format = fmt.Sprintf("// %s> %s\n", lg.prefix, format)
	}
	_, err := fmt.Printf(format, args...)
	return err
}

func (lg *LogGen) stepf(format string, args ...interface{}) {
	lg.step++
	format = fmt.Sprintf("step[%v]: %s", lg.step, format)
	lg.Logf(format, args...)
}

// Init prints a summary block describing the problem to be solved.
func (lg *LogGen) Init(desc string) {
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
	lg.Logf("Problem:")
	lg.Logf("  %s%v", strings.Repeat(" ", w-len(lg.Words[0])), string(lg.Words[0]))
	lg.Logf("+ %s%v", strings.Repeat(" ", w-len(lg.Words[1])), string(lg.Words[1]))
	lg.Logf("= %s%v", strings.Repeat(" ", w-len(lg.Words[2])), string(lg.Words[2]))
	lg.Logf("base: %v", lg.Base)
	lg.Logf("letters: %v", letters)
	lg.Logf("method: %s", desc)
	lg.Logf("")
}

// Fork changes the current prefix to be the cont string, while returing a new
// LogGen whose prefix is the alt string.  Additionally the prefixes have a
// level of indent that increases with repeated calls to Fork.
func (lg *LogGen) Fork(prob *PlanProblem, name, alt, cont string) SolutionGen {
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
	return &LogGen{
		PlanProblem: prob,
		prefix:      fmt.Sprintf("%s%s", strings.Repeat(" ", n+2), alt),
		step:        lg.step,
		branches:    lg.branches,
	}
}

// Fix prints a step log noting the fixed character.
func (lg *LogGen) Fix(c byte, v int) {
	lg.stepf("fix %v = %v", string(c), v)
}

// ComputeSum prints a step log noting the formula to be computed.
func (lg *LogGen) ComputeSum(col *Column) {
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

// ComputeFirstSummand prints a step log noting the formula to be computed.
func (lg *LogGen) ComputeFirstSummand(col *Column) {
	lg.computeSummand(col.Chars[0], col.Chars[1], col.Chars[2])
}

// ComputeSecondSummand prints a step log noting the formula to be computed.
func (lg *LogGen) ComputeSecondSummand(col *Column) {
	lg.computeSummand(col.Chars[1], col.Chars[0], col.Chars[2])
}

func (lg *LogGen) computeSummand(a, b, c byte) {
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

// ChooseRange prints a step log noting the range of futures to be explored.
func (lg *LogGen) ChooseRange(c byte, min, max int) {
	N := max - min
	R := lg.Base - len(lg.Known)
	if R < N {
		N = R
	}
	lg.branches = append(lg.branches, N)
	lg.stepf("choose %v (branch by %v)", string(c), N)
}

// CheckColumn prints a step log noting the column to be checked.
func (lg *LogGen) CheckColumn(col *Column, err error) {
	lg.stepf("check column: %s", col.Label())
}

// Verify prints a step log noting the verification phase.
func (lg *LogGen) Verify() {
	lg.stepf("verify")
}

// Check prints a step log noting the overall check.
func (lg *LogGen) Check(err error) {
	lg.stepf("check")
}

// Finish prints a step log noting the end of this future; in other words every
// path in the tree created by Fork will end in a Finish.
func (lg *LogGen) Finish() {
	lg.stepf("finish")
}

// Finalize prints final information once the entire plan is all done; this
// happens once all Fork branches have been Finished.
func (lg *LogGen) Finalize() {
	branches := 1
	for _, b := range lg.branches {
		branches *= b
	}

	lg.Logf("Total Branches: %v", branches)
}
