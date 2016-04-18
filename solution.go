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
	for i, word := range prob.words {
		pad := strings.Repeat(" ", len(prob.words[2])-len(word))
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

func (sol *solution) fork(v int) {
	other := sol.pool.Get()
	*other = *sol
	other.stepi = sol.stepi - 1
	other.ra = v
	sol.emit(other)
}
