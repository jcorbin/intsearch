package main

import "fmt"

type saveStep struct {
}

func (step saveStep) String() string {
	return fmt.Sprintf("save")
}

func (step saveStep) run(sol *solution) {
	sol.rb = sol.ra
}

type restoreStep struct {
}

func (step restoreStep) String() string {
	return fmt.Sprintf("restore")
}

func (step restoreStep) run(sol *solution) {
	sol.ra = sol.rb
}

type setStep int

func (v setStep) String() string {
	return fmt.Sprintf("set(%v)", int(v))
}

func (v setStep) run(sol *solution) {
	sol.ra = int(v)
}

type addStep byte

func (c addStep) String() string {
	return fmt.Sprintf("add(%s)", string(c))
}

func (c addStep) run(sol *solution) {
	sol.ra += sol.values[c]
}

type subStep byte

func (c subStep) String() string {
	return fmt.Sprintf("sub(%s)", string(c))
}

func (c subStep) run(sol *solution) {
	sol.ra -= sol.values[c]
}

type divStep int

func (v divStep) String() string {
	return fmt.Sprintf("div(%v)", int(v))
}

func (v divStep) run(sol *solution) {
	if sol.ra < 0 {
		sol.ra = -sol.ra / int(v)
	} else {
		sol.ra = sol.ra / int(v)
	}
}

type negateStep struct {
}

func (step negateStep) String() string {
	return fmt.Sprintf("negate")
}

func (step negateStep) run(sol *solution) {
	sol.ra = -sol.ra
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
	if sol.used[sol.ra] {
		sol.exit(errAlreadyUsed)
	}
	sol.values[c] = sol.ra
	sol.used[sol.ra] = true
}

type relJZStep int

func (o relJZStep) String() string {
	return fmt.Sprintf("jz(%+d)", int(o))
}

func (o relJZStep) run(sol *solution) {
	if sol.ra == 0 {
		sol.stepi += int(o)
	}
}

type relJNZStep int

func (o relJNZStep) String() string {
	return fmt.Sprintf("jnz(%+d)", int(o))
}

func (o relJNZStep) run(sol *solution) {
	if sol.ra != 0 {
		sol.stepi += int(o)
	}
}

type modStep int

func (v modStep) String() string {
	return fmt.Sprintf("mod(%v)", int(v))
}

func (v modStep) run(sol *solution) {
	sol.ra = (sol.ra + int(v)<<1) % int(v)
}

type forkUntilStep int

func (v forkUntilStep) String() string {
	return fmt.Sprintf("forkUntil(%v)", int(v))
}

func (v forkUntilStep) run(sol *solution) {
	if sol.ra < int(v) {
		sol.fork(sol.ra + 1)
	}
}
