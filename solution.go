package main

import (
	"fmt"
	"strings"
	"sync"
)

type solutionPool struct {
	sync.Pool
}

func (sp *solutionPool) Get() *solution {
	sol, _ := sp.Pool.Get().(*solution)
	if sol == nil {
		sol = &solution{}
	}
	sol.pool = sp
	return sol
}

func (sp *solutionPool) Put(sol *solution) {
	if sol.trace != nil {
		for i := range sol.trace {
			// sp.Put(sol.trace[i]) XXX useful?
			sol.trace[i] = nil
		}
		sol.trace = sol.trace[:0]
	}
	sp.Pool.Put(sol)
}

type solutionStep interface {
	run(sol *solution)
}

type labeledStep interface {
	labelName() string
	eraseLabel() []solutionStep
}

type resolvableStep interface {
	resolveLabels(labels map[string]int) solutionStep
}

type labelStep string

func (l labelStep) labelName() string {
	return string(l)
}

func (l labelStep) eraseLabel() []solutionStep {
	return nil
}

func (l labelStep) String() string {
	return fmt.Sprintf(":%s", string(l))
}

func (l labelStep) run(sol *solution) {
}

// extractLabels collects all labeledStep addresses from the list of steps
// passed.  After return every labeledStep, ls, has an entry in labels,
// labels[ls.labelName()] == addr, such that steps[addr] == ls.
func extractLabels(steps []solutionStep, labels map[string]int) map[string]int {
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
			labels[ls.labelName()] = addr
		}
	}
	return labels
}

// resolveLabels calls all step.resolveLabels methods for all resolvableSteps
// in steps. Each resolvableStep is replaced by any non-nil step its
// resolveLabels method returned.  If the passed labels map is nil, then
// extractLabels is called to build it.  Both the modified steps and labels map
// are returned.
func resolveLabels(steps []solutionStep, labels map[string]int) ([]solutionStep, map[string]int) {
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

// eraseLabels erases all labeledSteps in steps, updating passed labels values
// as appropriate.
//
// For each labeledStep, the step is replaced by the (maybe empty) list of
// steps returned by step.eraseLabel().
func eraseLabels(addr int, steps []solutionStep, parts [][]solutionStep, labels map[string]int) (int, [][]solutionStep, map[string]int) {
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
		parts = make([][]solutionStep, 0, 2*nl+1)
	}
	var prior int
	for i, step := range steps {
		if ls, ok := step.(labeledStep); ok {
			labels[ls.labelName()] = addr
			if head := steps[prior:i]; len(head) > 0 {
				parts = append(parts, head)
			}
			if replace := ls.eraseLabel(); len(replace) > 0 {
				parts = append(parts, replace)
				addr += len(replace)
			}
			prior = i + 1
		} else {
			addr++
		}
	}
	if tail := steps[prior:]; len(tail) > 0 {
		parts = append(parts, tail)
		addr += len(tail)
	}
	return addr, parts, labels
}

type solution struct {
	prob       *problem
	pool       *solutionPool
	emit       func(*solution)
	steps      []solutionStep
	stepi      int
	values     [256]int
	used       [256]bool
	ra, rb, rc int
	done       bool
	err        error
	trace      []*solution
}

func newSolution(prob *problem, steps []solutionStep, emit func(*solution)) *solution {
	sol := solution{
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

func (sol *solution) String() string {
	var step solutionStep
	if sol.stepi < len(sol.steps) {
		step = sol.steps[sol.stepi]
	}
	return fmt.Sprintf("ra:%v rb:%v rc:%v done:%v err:%v -- @%v %v",
		sol.ra, sol.rb, sol.rc,
		sol.done, sol.err,
		sol.stepi, step,
	)
}

func (sol *solution) printCheck(printf func(string, ...interface{})) {
	ns := sol.numbers()
	check := ns[0]+ns[1] == ns[2]
	printf("Check: %v", check)
	marks := []string{" ", "+", "="}
	rels := []string{"==", "==", "=="}
	if !check {
		rels[2] = "!="
	}
	for i, word := range sol.prob.words {
		pad := strings.Repeat(" ", len(sol.prob.words[2])-len(word))
		printf("  %s%s %s == %s%v", marks[i], pad, word, pad, ns[i])
	}
}

func (sol *solution) numbers() [3]int {
	var ns [3]int
	base := sol.prob.base
	for i, word := range sol.prob.words {
		n := 0
		for _, c := range word {
			n = n*base + sol.values[c]
		}
		ns[i] = n
	}
	return ns
}

func (sol *solution) letterMapping() string {
	parts := make([]string, 0, len(sol.prob.letterSet))
	for _, l := range sol.prob.sortedLetters() {
		c := l[0]
		v := sol.values[c]
		if v >= 0 && sol.used[v] {
			parts = append(parts, fmt.Sprintf("%v:%v", l, v))
		}
	}
	return strings.Join(parts, " ")
}

func (sol *solution) step() bool {
	step := sol.steps[sol.stepi]
	sol.stepi++
	step.run(sol)
	return !sol.done
}

func (sol *solution) exit(err error) {
	sol.done = true
	sol.err = err
	if err != nil {
		sol.stepi--
	}
}

func (sol *solution) copy() *solution {
	other := sol.pool.Get()
	*other = *sol
	return other
}
