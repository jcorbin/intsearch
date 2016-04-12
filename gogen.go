package main

import (
	"errors"
	"fmt"
)

// TODO: currently broken, probably around incorrect inter-column carry
// computation after a negated solve

var (
	errAlreadyUsed = errors.New("value already used")
	errCheckFailed = errors.New("check failed")
)

type goGen struct {
	steps []solutionStep
}

func (gg *goGen) obsAfter() *afterGen {
	i := 0
	return &afterGen{func(prob *problem) {
		j := i
		for ; j < len(gg.steps); j++ {
			fmt.Printf("%v: %v\n", j, gg.steps[j])
		}
		fmt.Println()
		i = j
	}}
}

type saveStep struct {
}

func (step saveStep) String() string {
	return fmt.Sprintf("save")
}

func (step saveStep) run(sol *solution) {
	sol.save = sol.carry
}

type restoreStep struct {
}

func (step restoreStep) String() string {
	return fmt.Sprintf("restore")
}

func (step restoreStep) run(sol *solution) {
	sol.carry = sol.save
}

type setStep int

func (v setStep) String() string {
	return fmt.Sprintf("set(%v)", int(v))
}

func (v setStep) run(sol *solution) {
	sol.carry = int(v)
}

type addStep byte

func (c addStep) String() string {
	return fmt.Sprintf("add(%s)", string(c))
}

func (c addStep) run(sol *solution) {
	sol.carry += sol.values[c]
}

type divStep int

func (v divStep) String() string {
	return fmt.Sprintf("div(%v)", int(v))
}

func (v divStep) run(sol *solution) {
	if sol.carry < 0 {
		sol.carry = -sol.carry / int(v)
	} else {
		sol.carry = sol.carry / int(v)
	}
}

type negateStep struct {
}

func (step negateStep) String() string {
	return fmt.Sprintf("negate")
}

func (step negateStep) run(sol *solution) {
	sol.carry = -sol.carry
}

type exitStep struct {
	err error
}

func (step exitStep) String() string {
	return fmt.Sprintf("exit(%v)", step.err)
}

func (step exitStep) run(sol *solution) {
	sol.exit(step.err)
}

type storeStep byte

func (c storeStep) String() string {
	return fmt.Sprintf("store(%s)", string(c))
}

func (c storeStep) run(sol *solution) {
	if sol.used[sol.carry] {
		sol.exit(errAlreadyUsed)
	}
	sol.values[c] = sol.carry
	sol.used[sol.carry] = true
}

type checkStep byte

func (c checkStep) String() string {
	return fmt.Sprintf("check(%s)", string(c))
}

func (c checkStep) run(sol *solution) {
	if sol.values[c] != sol.carry {
		sol.exit(errCheckFailed)
	}
}

type modStep int

func (v modStep) String() string {
	return fmt.Sprintf("mod(%v)", int(v))
}

func (v modStep) run(sol *solution) {
	sol.carry = (sol.carry + int(v)<<1) % int(v)
}

type forkUntilStep int

func (v forkUntilStep) String() string {
	return fmt.Sprintf("forkUntil(%v)", int(v))
}

func (v forkUntilStep) run(sol *solution) {
	if sol.carry < int(v) {
		sol.fork(sol.carry + 1)
	}
}

func (gg *goGen) init(prob *problem, desc string) {
}

func (gg *goGen) fix(prob *problem, c byte, v int) {
	gg.steps = append(gg.steps, setStep(v))
	gg.steps = append(gg.steps, storeStep(c))
}

func (gg *goGen) initColumn(prob *problem, cx [3]byte, numKnown, numUnknown int) {
}

func (gg *goGen) computeSum(prob *problem, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for c:
	//   c = carry + a + b (mod base)
	gg.steps = append(gg.steps, saveStep{})
	if a != 0 {
		gg.steps = append(gg.steps, addStep(a))
	}
	if b != 0 {
		gg.steps = append(gg.steps, addStep(b))
	}
	gg.steps = append(gg.steps, modStep(prob.base))
	gg.steps = append(gg.steps, storeStep(c))
	gg.steps = append(gg.steps, restoreStep{})
}

func (gg *goGen) computeSummand(prob *problem, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for a:
	//   a = c - b - carry (mod base)
	gg.steps = append(gg.steps, saveStep{})
	if c != 0 {
		gg.steps = append(gg.steps, addStep(c))
	}
	if b != 0 {
		gg.steps = append(gg.steps, addStep(b))
	}
	gg.steps = append(gg.steps, negateStep{})
	gg.steps = append(gg.steps, modStep(prob.base))
	gg.steps = append(gg.steps, storeStep(a))
	gg.steps = append(gg.steps, restoreStep{})
}

func (gg *goGen) computeCarry(prob *problem, c1, c2 byte) {
	if c1 != 0 {
		gg.steps = append(gg.steps, addStep(c1))
	}
	if c2 != 0 {
		gg.steps = append(gg.steps, addStep(c2))
	}
	gg.steps = append(gg.steps, divStep(prob.base))
}

func (gg *goGen) choose(prob *problem, c byte) {
	gg.steps = append(gg.steps, saveStep{})
	gg.steps = append(gg.steps, setStep(0))
	gg.steps = append(gg.steps, forkUntilStep(prob.base-1))
	gg.steps = append(gg.steps, storeStep(c))
	gg.steps = append(gg.steps, restoreStep{})
}

func (gg *goGen) checkFinal(prob *problem, c byte, c1, c2 byte) {
	gg.steps = append(gg.steps, storeStep(c))
}

func (gg *goGen) finish(prob *problem) {
	gg.steps = append(gg.steps, exitStep{nil})
}
