package main

import (
	"fmt"
	"strings"
)

type solutionStep interface {
	run(sol *solution)
}

type solution struct {
	prob   *problem
	emit   func(*solution)
	steps  []solutionStep
	stepi  int
	values [256]int
	used   [256]bool
	carry  int
	save   int
	done   bool
	err    error
}

func newSolution(prob *problem, steps []solutionStep, emit func(*solution)) *solution {
	sol := solution{
		prob:  prob,
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
	return fmt.Sprintf("carry:%v save:%v done:%v err:%v -- @%v %v",
		sol.carry, sol.save,
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
	return &solution{
		prob:   sol.prob,
		emit:   sol.emit,
		steps:  sol.steps,
		stepi:  sol.stepi,
		values: sol.values,
		used:   sol.used,
		carry:  sol.carry,
		save:   sol.save,
		done:   sol.done,
		err:    sol.err,
	}
}

func (sol *solution) fork(v int) {
	sol.emit(&solution{
		prob:   sol.prob,
		emit:   sol.emit,
		steps:  sol.steps,
		stepi:  sol.stepi - 1,
		values: sol.values,
		used:   sol.used,
		carry:  v,
		save:   sol.save,
		done:   sol.done,
		err:    sol.err,
	})
}
