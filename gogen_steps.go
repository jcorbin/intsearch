package main

import "fmt"

type saveStep struct{}
type restoreStep struct{}

func (step saveStep) String() string    { return fmt.Sprintf("save") }
func (step restoreStep) String() string { return fmt.Sprintf("restore") }

func (step saveStep) run(sol *solution)    { sol.rb = sol.ra }
func (step restoreStep) run(sol *solution) { sol.ra = sol.rb }

type setStep int

func (step setStep) String() string    { return fmt.Sprintf("set(%v)", int(step)) }
func (step setStep) run(sol *solution) { sol.ra = int(step) }

type addStep byte
type subStep byte
type divStep int
type modStep int

func (step addStep) String() string { return fmt.Sprintf("add(%s)", string(step)) }
func (step subStep) String() string { return fmt.Sprintf("sub(%s)", string(step)) }
func (step divStep) String() string { return fmt.Sprintf("div(%v)", int(step)) }
func (step modStep) String() string { return fmt.Sprintf("mod(%v)", int(step)) }

func (step addStep) run(sol *solution) { sol.ra += sol.values[step] }
func (step subStep) run(sol *solution) { sol.ra -= sol.values[step] }
func (step divStep) run(sol *solution) {
	if sol.ra < 0 {
		sol.ra = -sol.ra / int(step)
	} else {
		sol.ra = sol.ra / int(step)
	}
}
func (step modStep) run(sol *solution) { sol.ra = (sol.ra + int(step)<<1) % int(step) }

type negateStep struct{}

func (step negateStep) String() string    { return fmt.Sprintf("negate") }
func (step negateStep) run(sol *solution) { sol.ra = -sol.ra }

type exitStep struct{ err error }

func (step exitStep) String() string    { return fmt.Sprintf("exit(%v)", step.err) }
func (step exitStep) run(sol *solution) { sol.exit(step.err) }

type storeStep byte

func (c storeStep) String() string { return fmt.Sprintf("store(%s)", string(c)) }
func (c storeStep) run(sol *solution) {
	if sol.used[sol.ra] {
		sol.exit(errAlreadyUsed)
	}
	sol.values[c] = sol.ra
	sol.used[sol.ra] = true
}

type relJZStep int
type relJNZStep int

func (step relJZStep) String() string  { return fmt.Sprintf("jz(%+d)", int(step)) }
func (step relJNZStep) String() string { return fmt.Sprintf("jnz(%+d)", int(step)) }

func (step relJZStep) run(sol *solution) {
	if sol.ra == 0 {
		sol.stepi += int(step)
	}
}
func (step relJNZStep) run(sol *solution) {
	if sol.ra != 0 {
		sol.stepi += int(step)
	}
}

type forkUntilStep int

func (step forkUntilStep) String() string { return fmt.Sprintf("forkUntil(%v)", int(step)) }
func (step forkUntilStep) run(sol *solution) {
	if sol.ra < int(step) {
		sol.fork(sol.ra + 1)
	}
}
