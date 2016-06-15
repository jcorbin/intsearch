package runnable

import (
	"fmt"
	"sync"

	"github.com/jcorbin/intsearch/word"
)

type solutionPool struct {
	sync.Pool
}

func (sp *solutionPool) Get() *Solution {
	sol, _ := sp.Pool.Get().(*Solution)
	if sol == nil {
		sol = &Solution{}
	}
	sol.pool = sp
	return sol
}

func (sp *solutionPool) Put(sol *Solution) {
	sp.Pool.Put(sol)
}

// Step is a single runnable step that will make progress on a solution.
type Step interface {
	run(sol *Solution)
}

type labeledStep interface {
	labelName() string
}

type expandableStep interface {
	expandStep(
		addr int,
		parts [][]Step,
		labels map[string]int,
		annotate annoFunc,
	) (int, [][]Step, map[string]int)
}

type annotatedStep interface {
	annotate() string
}

type resolvableStep interface {
	resolveLabels(labels map[string]int) Step
}

type labelStep string

func (l labelStep) labelName() string {
	return string(l)
}

func (l labelStep) expandStep(
	addr int,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]Step, map[string]int) {
	if annotate != nil {
		annotate(addr, l.String())
	}
	return addr, parts, labels
}

func (l labelStep) String() string {
	return fmt.Sprintf(":%s", string(l))
}

func (l labelStep) run(sol *Solution) {
}

// extractLabels collects all labeledStep addresses from the list of steps
// passed.  After return every labeledStep, ls, has an entry in labels,
// labels[ls.labelName()] == addr, such that steps[addr] == ls.
func extractLabels(steps []Step, labels map[string]int) map[string]int {
	n := 0
	for _, step := range steps {
		if _, ok := step.(labeledStep); ok {
			n++
		}
	}
	if n == 0 {
		if labels == nil {
			labels = make(map[string]int)
		}
		return labels
	}
	if labels == nil {
		labels = make(map[string]int, n)
	} else {
		nl := make(map[string]int, len(labels)+n)
		for k, v := range labels {
			nl[k] = v
		}
		labels = nl
	}
	for addr, step := range steps {
		if ls, ok := step.(labeledStep); ok {
			if name := ls.labelName(); name != "" {
				labels[name] = addr
			}
		}
	}
	return labels
}

// resolveLabels calls all step.resolveLabels methods for all resolvableSteps
// in steps. Each resolvableStep is replaced by any non-nil step its
// resolveLabels method returned.  If the passed labels map is nil, then
// extractLabels is called to build it.  Both the modified steps and labels map
// are returned.
func resolveLabels(steps []Step, labels map[string]int) ([]Step, map[string]int) {
	if labels == nil {
		labels = extractLabels(steps, nil)
	}
	for addr, step := range steps {
		if rs, ok := step.(resolvableStep); ok {
			if step := rs.resolveLabels(labels); step != nil {
				steps[addr] = step
			}
		}
	}
	return steps, labels
}

type annoFunc func(addr int, annos ...string)

type stepExpander func(
	es expandableStep,
	addr int,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]Step, map[string]int)

// expandSteps expands all expandableSteps.
//
// For each expandableStep, step.expandStep(addr, parts, labels) is called;
// this method may append zero or more new parts, and should add any labels
// contained in those parts to the labels map.  The returned addr must be the
// passed addr plus the total step length of all newly added parts.  The, maybe
// modified, parts and labels must be returned.
//
// Implementations are expected to call expandSteps on any newly added parts
// that need it; recursive step expansion is not provided by expandSteps.
func expandSteps(
	addr int,
	steps []Step,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]Step, map[string]int) {
	return actuallyExpandSteps(addr, steps, parts, labels, annotate, nil)
}

func debugExpandSteps(
	addr int,
	steps []Step,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]Step, map[string]int) {
	fmt.Println()
	fmt.Printf("// expanding steps @%d\n", addr)
	for i, step := range steps {
		fmt.Printf("%d: %v\n", addr+i, step)
	}

	startAddr := addr
	startPartsLen := len(parts)

	addr, parts, labels = actuallyExpandSteps(
		addr, steps, parts, labels, annotate,
		mustExpandStepSanely)

	fmt.Println()
	fmt.Printf("// expanded parts @%d\n", addr)
	for i, part := range parts[startPartsLen:] {
		fmt.Printf("// part %d\n", i)
		for _, step := range part {
			if rs, ok := step.(resolvableStep); ok {
				os := rs.resolveLabels(labels)
				fmt.Printf("%d: %v // %v\n", startAddr, step, os)
			} else {
				fmt.Printf("%d: %v\n", startAddr, step)
			}
			startAddr++
		}
	}

	return addr, parts, labels
}

func actuallyExpandSteps(
	addr int,
	steps []Step,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
	expand stepExpander,
) (int, [][]Step, map[string]int) {
	if parts == nil {
		nl := len(labels)
		if labels == nil {
			for _, step := range steps {
				if _, ok := step.(labeledStep); ok {
					nl++
				}
			}
			labels = make(map[string]int, nl)
		}
		parts = make([][]Step, 0, 2*nl+1)
	}
	var prior int
	for i, step := range steps {
		if ls, ok := step.(labeledStep); ok {
			if name := ls.labelName(); name != "" {
				labels[name] = addr
			}
		}
		if es, ok := step.(expandableStep); ok {
			if head := steps[prior:i]; len(head) > 0 {
				parts = append(parts, head)
			}
			if expand != nil {
				addr, parts, labels = expand(es, addr, parts, labels, annotate)
			} else {
				addr, parts, labels = es.expandStep(addr, parts, labels, annotate)
			}
			prior = i + 1
			continue
		}
		if annotate != nil {
			if as, ok := step.(annotatedStep); ok {
				annotate(addr, as.annotate())
			}
		}
		addr++
	}
	if tail := steps[prior:]; len(tail) > 0 {
		parts = append(parts, tail)
	}
	return addr, parts, labels
}

func mustExpandStepSanely(
	es expandableStep,
	addr int,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (
	newAddr int,
	newParts [][]Step,
	newLabels map[string]int,
) {
	newAddr, newParts, newLabels = es.expandStep(addr, parts, labels, annotate)
	diff := newAddr - addr
	for _, part := range newParts[len(parts):] {
		diff -= len(part)
	}
	if diff != 0 {
		panic(fmt.Sprintf(
			"failed to correctly expand steps @%d: expanded addr(%d) is off by %d",
			addr, newAddr, diff))
	}
	return
}

// Solution is a problem, a built program to solve the problem, and the state
// for that program.  Solutions are only created by StepsGen.SearchInit or as
// copies of an existing solution.
type Solution struct {
	prob       *word.Problem
	pool       *solutionPool
	emit       func(*Solution)
	steps      []Step
	stepi      int
	values     [256]int
	used       [256]bool
	ra, rb, rc int
	done       bool
	err        error
}

func newSolution(prob *word.Problem, steps []Step, emit func(*Solution)) *Solution {
	sol := Solution{
		prob:  prob,
		pool:  &solutionPool{},
		emit:  emit,
		steps: steps,
	}
	for i := 0; i < 256; i++ {
		sol.values[i] = -1
	}
	return &sol
}

// Problem returns the associated problem.
func (sol *Solution) Problem() *word.Problem {
	return sol.prob
}

// ValueOf returns the value of a single letter, and whether or not it is
// actually known.
func (sol *Solution) ValueOf(c byte) (v int, known bool) {
	v = sol.values[c]
	known = v >= 0 && sol.used[v]
	return
}

// Check returns any solution error, or word.ErrSolutionNotDone if not done.
func (sol *Solution) Check() error {
	if sol.err != nil {
		return sol.err
	}
	if !sol.done {
		return word.ErrSolutionNotDone
	}
	return nil
}

// Dump dumps the solution to a formatter.
func (sol *Solution) Dump(logf func(string, ...interface{})) {
	var last Step
	if sol.stepi > 0 {
		last = sol.steps[sol.stepi-1]
	}
	logf("... %v", sol)
	if isStoreStep(last) {
		logf("... %s", word.SolutionMapping(sol))
	}
}

// Err returns any execution error.
func (sol *Solution) Err() error {
	return sol.err
}

func (sol *Solution) String() string {
	var step Step
	if sol.stepi < len(sol.steps) {
		step = sol.steps[sol.stepi]
	}
	return fmt.Sprintf("ra:%v rb:%v rc:%v done:%v err:%v -- @%v %v",
		sol.ra, sol.rb, sol.rc,
		sol.done, sol.err,
		sol.stepi, step,
	)
}

// PaddedString returns a white-spaced padded version of .String().
func (sol *Solution) PaddedString() string {
	var step Step
	if sol.stepi < len(sol.steps) {
		step = sol.steps[sol.stepi]
	}
	return fmt.Sprintf(
		"ra:%-3v rb:%-3v rc:%-3v done:%v err:%v -- @%-3v %-20v",
		sol.ra, sol.rb, sol.rc,
		sol.done, sol.err,
		sol.stepi, step,
	)
}

// Numbers returns 3 numbers computed for the solution (as determined by the
// letter mapping).
func (sol *Solution) Numbers() [3]int {
	var ns [3]int
	base := sol.prob.Base
	for i, word := range sol.prob.Words {
		n := 0
		for _, c := range word {
			n = n*base + sol.values[c]
		}
		ns[i] = n
	}
	return ns
}

// Step runs a single step against the solution, and returns true if the
// solution has not terminated (should be steped again).
func (sol *Solution) Step() bool {
	if sol.done {
		return false
	}
	step := sol.steps[sol.stepi]
	sol.stepi++
	step.run(sol)
	return !sol.done
}

func (sol *Solution) exit(err error) {
	sol.done = true
	sol.err = err
	if err != nil {
		sol.stepi--
	}
}

func (sol *Solution) copy() *Solution {
	other := sol.pool.Get()
	*other = *sol
	return other
}
