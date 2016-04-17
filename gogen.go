package main

import (
	"errors"
	"fmt"
)

// TODO: currently we find the correct solution, but also find ~24 incorrect
// "solution"s; either there are bugs in the current computation logic, or we
// just need more checks

var (
	errAlreadyUsed  = errors.New("value already used")
	errCheckFailed  = errors.New("check failed")
	errVerifyFailed = errors.New("verify failed")
)

type goGen struct {
	steps      []solutionStep
	verified   bool
	carrySaved bool
	carryValid bool
}

func (gg *goGen) obsAfter() *afterGen {
	i := 0
	return &afterGen{func(prob *problem) {
		j := i
		for ; j < len(gg.steps); j++ {
			fmt.Printf("%v: %v\n", j, gg.steps[j])
		}
		if j > i {
			fmt.Println()
			i = j
		}
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

type subStep byte

func (c subStep) String() string {
	return fmt.Sprintf("sub(%s)", string(c))
}

func (c subStep) run(sol *solution) {
	sol.carry -= sol.values[c]
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

type relJZStep int

func (o relJZStep) String() string {
	return fmt.Sprintf("jz(%+d)", int(o))
}

func (o relJZStep) run(sol *solution) {
	if sol.carry == 0 {
		sol.stepi += int(o)
	}
}

type relJNZStep int

func (o relJNZStep) String() string {
	return fmt.Sprintf("jnz(%+d)", int(o))
}

func (o relJNZStep) run(sol *solution) {
	if sol.carry != 0 {
		sol.stepi += int(o)
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
	gg.steps = append(gg.steps, setStep(0))
	gg.carrySaved = false
	gg.carryValid = true
}

func (gg *goGen) fix(prob *problem, c byte, v int) {
	gg.steps = append(gg.steps, setStep(v))
	gg.steps = append(gg.steps, storeStep(c))
}

func (gg *goGen) initColumn(prob *problem, cx [3]byte, numKnown, numUnknown int) {
}

func (gg *goGen) saveCarry(prob *problem) {
	if !gg.carrySaved {
		if !gg.carryValid {
			panic("no valid carry to save")
		}
		gg.steps = append(gg.steps, saveStep{})
		gg.carrySaved = true
	}
}

func (gg *goGen) restoreCarry(prob *problem) {
	if !gg.carryValid {
		if !gg.carrySaved {
			panic("no saved carry to restore")
		}
		gg.steps = append(gg.steps, restoreStep{})
		gg.carryValid = true
	}
}

func (gg *goGen) computeSum(prob *problem, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for c:
	//   c = carry + a + b (mod base)
	gg.saveCarry(prob)
	gg.carryValid = false
	if a != 0 {
		gg.steps = append(gg.steps, addStep(a))
	}
	if b != 0 {
		gg.steps = append(gg.steps, addStep(b))
	}
	gg.steps = append(gg.steps, modStep(prob.base))
	gg.steps = append(gg.steps, storeStep(c))
	if c == prob.words[0][0] || c == prob.words[1][0] || c == prob.words[2][0] {
		gg.steps = append(gg.steps, relJNZStep(1))
		gg.steps = append(gg.steps, exitStep{errCheckFailed})
	}
	gg.restoreCarry(prob)
}

func (gg *goGen) computeSummand(prob *problem, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for a:
	//   a = c - b - carry (mod base)
	gg.saveCarry(prob)
	gg.carryValid = false
	gg.steps = append(gg.steps, negateStep{})
	if c != 0 {
		gg.steps = append(gg.steps, addStep(c))
	}
	if b != 0 {
		gg.steps = append(gg.steps, subStep(b))
	}
	gg.steps = append(gg.steps, modStep(prob.base))
	gg.steps = append(gg.steps, storeStep(a))
	if a == prob.words[0][0] || a == prob.words[1][0] || a == prob.words[2][0] {
		gg.steps = append(gg.steps, relJNZStep(1))
		gg.steps = append(gg.steps, exitStep{errCheckFailed})
	}
	gg.restoreCarry(prob)
}

func (gg *goGen) computeCarry(prob *problem, c1, c2 byte) {
	if c1 != 0 {
		gg.steps = append(gg.steps, addStep(c1))
	}
	if c2 != 0 {
		gg.steps = append(gg.steps, addStep(c2))
	}
	gg.steps = append(gg.steps, divStep(prob.base))
	gg.carryValid = true
	gg.carrySaved = false
}

func (gg *goGen) choose(prob *problem, c byte) {
	gg.saveCarry(prob)
	gg.carryValid = false
	if c == prob.words[0][0] || c == prob.words[1][0] || c == prob.words[2][0] {
		gg.steps = append(gg.steps, setStep(1))
	} else {
		gg.steps = append(gg.steps, setStep(0))
	}
	gg.steps = append(gg.steps, forkUntilStep(prob.base-1))
	gg.steps = append(gg.steps, storeStep(c))
	gg.restoreCarry(prob)
}

func (gg *goGen) checkFinal(prob *problem, c byte, c1, c2 byte) {
	gg.steps = append(gg.steps, subStep(c))
	gg.steps = append(gg.steps, relJZStep(1))
	gg.steps = append(gg.steps, exitStep{errCheckFailed})
}

func (gg *goGen) verify(prob *problem) {
	gg.steps = append(gg.steps, setStep(0))
	prob.eachColumn(func(cx [3]byte) {
		if cx[0] != 0 {
			gg.steps = append(gg.steps, addStep(cx[0]))
		}
		if cx[1] != 0 {
			gg.steps = append(gg.steps, addStep(cx[1]))
		}
		gg.steps = append(gg.steps, saveStep{})
		gg.steps = append(gg.steps, modStep(prob.base))
		gg.steps = append(gg.steps, subStep(cx[2]))
		gg.steps = append(gg.steps, relJZStep(1))
		gg.steps = append(gg.steps, exitStep{errVerifyFailed})
		gg.steps = append(gg.steps, restoreStep{})
		gg.steps = append(gg.steps, divStep(prob.base))
	})
	gg.steps = append(gg.steps, relJZStep(1))
	gg.steps = append(gg.steps, exitStep{errVerifyFailed})
}

func (gg *goGen) finish(prob *problem) {
	if gg.verified {
		gg.verify(prob)
	}
	gg.steps = append(gg.steps, exitStep{nil})
}
