package word

import (
	"errors"
	"fmt"
	"strings"
)

// ErrSolutionNotDone is returned by a Solution.Check() if it is not yet
// complete.
var ErrSolutionNotDone = errors.New("solution not complete")

// VerifyError is the error returned if final verification fails.
type VerifyError string

func (ve VerifyError) Error() string {
	return fmt.Sprintf("verify failed: %s", string(ve))
}

// Solution is implemented by all concrete plan solutions.
type Solution interface {
	Problem() *Problem
	ValueOf(byte) (int, bool)
	Check() error
	Dump(logf func(string, ...interface{}))
	Trace() []Solution
}

// SolutionMapping returns a string describing a solution's letter
// mapping like "x:1 y:2 z:3".
func SolutionMapping(sol Solution) string {
	prob := sol.Problem()
	parts := make([]string, 0, len(prob.Letters))
	for _, c := range prob.SortedLetters() {
		if v, known := sol.ValueOf(c); known {
			parts = append(parts, fmt.Sprintf("%s:%v", string(c), v))
		}
	}
	return strings.Join(parts, " ")
}

// SolutionNumbers returns the 3 numbers computed by the solution (as determined by the
// letter mapping).
func SolutionNumbers(sol Solution) (ns [3]int) {
	prob := sol.Problem()
	base := prob.Base
	for i, word := range prob.Words {
		n := 0
		for _, c := range word {
			v, _ := sol.ValueOf(c)
			n = n*base + v
		}
		ns[i] = n
	}
	return
}

// SolutionCheck prints a simple double check of the solution.
func SolutionCheck(sol Solution, printf func(string, ...interface{})) {
	prob := sol.Problem()
	ns := SolutionNumbers(sol)
	check := ns[0]+ns[1] == ns[2]
	printf("Check: %v", check)
	marks := []string{" ", "+", "="}
	rels := []string{"==", "==", "=="}
	if !check {
		rels[2] = "!="
	}
	for i, word := range prob.Words {
		pad := strings.Repeat(" ", len(prob.Words[2])-len(word))
		printf("  %s%s %s == %s%v", marks[i], pad, word, pad, ns[i])
	}
}
