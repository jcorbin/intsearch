package main

import "fmt"

type setABStep struct{}
type setACStep struct{}
type setBAStep struct{}
type setBCStep struct{}
type setCAStep struct{}
type setCBStep struct{}

func (step setABStep) String() string { return fmt.Sprintf("ra = rb") }
func (step setACStep) String() string { return fmt.Sprintf("ra = rc") }
func (step setBAStep) String() string { return fmt.Sprintf("rb = ra") }
func (step setBCStep) String() string { return fmt.Sprintf("rb = rc") }
func (step setCAStep) String() string { return fmt.Sprintf("rc = ra") }
func (step setCBStep) String() string { return fmt.Sprintf("rc = rb") }

func (step setABStep) run(sol *solution) { sol.ra = sol.rb }
func (step setACStep) run(sol *solution) { sol.ra = sol.rc }
func (step setBAStep) run(sol *solution) { sol.rb = sol.ra }
func (step setBCStep) run(sol *solution) { sol.rb = sol.rc }
func (step setCAStep) run(sol *solution) { sol.rc = sol.ra }
func (step setCBStep) run(sol *solution) { sol.rc = sol.rb }

type setAStep int

func (step setAStep) String() string    { return fmt.Sprintf("ra = %v", int(step)) }
func (step setAStep) run(sol *solution) { sol.ra = int(step) }

type ltStep int
type lteStep int
type eqStep int
type gteStep int
type gtStep int

func (step ltStep) String() string  { return fmt.Sprintf("lt(%v)", int(step)) }
func (step lteStep) String() string { return fmt.Sprintf("lte(%v)", int(step)) }
func (step eqStep) String() string  { return fmt.Sprintf("eq(%v)", int(step)) }
func (step gteStep) String() string { return fmt.Sprintf("gte(%v)", int(step)) }
func (step gtStep) String() string  { return fmt.Sprintf("gt(%v)", int(step)) }

func (step ltStep) run(sol *solution) {
	if sol.ra < int(step) {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}
func (step lteStep) run(sol *solution) {
	if sol.ra <= int(step) {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}
func (step eqStep) run(sol *solution) {
	if sol.ra <= int(step) {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}
func (step gteStep) run(sol *solution) {
	if sol.ra >= int(step) {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}
func (step gtStep) run(sol *solution) {
	if sol.ra > int(step) {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}

type addValueStep byte
type subValueStep byte
type addStep int
type subStep int
type divStep int
type modStep int

func (step addValueStep) String() string { return fmt.Sprintf("add($%s)", string(step)) }
func (step subValueStep) String() string { return fmt.Sprintf("sub($%s)", string(step)) }
func (step addStep) String() string      { return fmt.Sprintf("add(%+d)", int(step)) }
func (step subStep) String() string      { return fmt.Sprintf("sub(%+d)", int(step)) }
func (step divStep) String() string      { return fmt.Sprintf("div(%v)", int(step)) }
func (step modStep) String() string      { return fmt.Sprintf("mod(%v)", int(step)) }

func (step addValueStep) run(sol *solution) { sol.ra += sol.values[step] }
func (step subValueStep) run(sol *solution) { sol.ra -= sol.values[step] }
func (step addStep) run(sol *solution)      { sol.ra += int(step) }
func (step subStep) run(sol *solution)      { sol.ra -= int(step) }
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

type jmpStep int

func (n jmpStep) String() string    { return fmt.Sprintf("jmp(%d)", int(n)) }
func (n jmpStep) run(sol *solution) { sol.stepi = int(n) }

type jzStep int
type jnzStep int

func (step jzStep) String() string  { return fmt.Sprintf("jz(%d)", int(step)) }
func (step jnzStep) String() string { return fmt.Sprintf("jnz(%d)", int(step)) }

func (step jzStep) run(sol *solution) {
	if sol.ra == 0 {
		sol.stepi = int(step)
	}
}
func (step jnzStep) run(sol *solution) {
	if sol.ra != 0 {
		sol.stepi = int(step)
	}
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
		child := sol.copy()
		child.stepi = sol.stepi - 1
		child.ra = sol.ra + 1
		sol.emit(child)
	}
}
