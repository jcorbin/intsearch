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
type setBStep int
type setCStep int

func (step setAStep) String() string    { return fmt.Sprintf("ra = %v", int(step)) }
func (step setBStep) String() string    { return fmt.Sprintf("rb = %v", int(step)) }
func (step setCStep) String() string    { return fmt.Sprintf("rc = %v", int(step)) }
func (step setAStep) run(sol *solution) { sol.ra = int(step) }
func (step setBStep) run(sol *solution) { sol.rb = int(step) }
func (step setCStep) run(sol *solution) { sol.rc = int(step) }

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

type negateStep struct{}
type addRegBStep struct{}
type addRegCStep struct{}
type subRegBStep struct{}
type subRegCStep struct{}
type addValueStep byte
type subValueStep byte
type addStep int
type subStep int
type modStep int
type divStep int

func (step negateStep) String() string   { return fmt.Sprintf("negate") }
func (step addRegBStep) String() string  { return "add %b" }
func (step addRegCStep) String() string  { return "add %c" }
func (step subRegBStep) String() string  { return "sub %b" }
func (step subRegCStep) String() string  { return "sub %c" }
func (step addValueStep) String() string { return fmt.Sprintf("add $%s", string(step)) }
func (step subValueStep) String() string { return fmt.Sprintf("sub $%s", string(step)) }
func (step addStep) String() string      { return fmt.Sprintf("add %+d", int(step)) }
func (step subStep) String() string      { return fmt.Sprintf("sub %+d", int(step)) }
func (step modStep) String() string      { return fmt.Sprintf("mod %v", int(step)) }
func (step divStep) String() string      { return fmt.Sprintf("div %v", int(step)) }

func (step negateStep) run(sol *solution)   { sol.ra = -sol.ra }
func (step addRegBStep) run(sol *solution)  { sol.ra += sol.rb }
func (step addRegCStep) run(sol *solution)  { sol.ra += sol.rc }
func (step subRegBStep) run(sol *solution)  { sol.ra -= sol.rb }
func (step subRegCStep) run(sol *solution)  { sol.ra -= sol.rc }
func (step addValueStep) run(sol *solution) { sol.ra += sol.values[step] }
func (step subValueStep) run(sol *solution) { sol.ra -= sol.values[step] }
func (step addStep) run(sol *solution)      { sol.ra += int(step) }
func (step subStep) run(sol *solution)      { sol.ra -= int(step) }
func (step modStep) run(sol *solution)      { sol.ra = (sol.ra + int(step)<<1) % int(step) }
func (step divStep) run(sol *solution) {
	if sol.ra < 0 {
		sol.ra = -sol.ra / int(step)
	} else {
		sol.ra = sol.ra / int(step)
	}
}

type exitStep struct{ err error }

func (step exitStep) String() string    { return fmt.Sprintf("exit(%v)", step.err) }
func (step exitStep) run(sol *solution) { sol.exit(step.err) }

type isUsedStep struct{}

func (step isUsedStep) String() string { return fmt.Sprintf("used?") }
func (step isUsedStep) run(sol *solution) {
	if sol.used[sol.ra] {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}

type storeStep byte
type loadStep byte

func (c storeStep) String() string { return fmt.Sprintf("store(%s)", string(c)) }
func (c loadStep) String() string  { return fmt.Sprintf("load(%s)", string(c)) }
func (c storeStep) run(sol *solution) {
	// TODO: drop guard, program can now use isUsedStep to guarantee this never happens
	if sol.used[sol.ra] {
		sol.exit(errAlreadyUsed)
	}
	sol.values[c] = sol.ra
	sol.used[sol.ra] = true
}
func (c loadStep) run(sol *solution) {
	sol.ra = sol.values[c]
}

type jmpStep int
type jzStep int
type jnzStep int
type relJMPStep int
type relJZStep int
type relJNZStep int
type labelJmpStep string
type labelJZStep string
type labelJNZStep string
type forkStep int
type relForkStep int
type labelForkStep string

func (step jmpStep) String() string       { return fmt.Sprintf("jmp @%d", int(step)) }
func (step jzStep) String() string        { return fmt.Sprintf("jz @%d", int(step)) }
func (step jnzStep) String() string       { return fmt.Sprintf("jnz @%d", int(step)) }
func (step relJMPStep) String() string    { return fmt.Sprintf("jmp %+d", int(step)) }
func (step relJZStep) String() string     { return fmt.Sprintf("jz %+d", int(step)) }
func (step relJNZStep) String() string    { return fmt.Sprintf("jnz %+d", int(step)) }
func (step labelJmpStep) String() string  { return fmt.Sprintf("jmp :%s", string(step)) }
func (step labelJZStep) String() string   { return fmt.Sprintf("jz :%s", string(step)) }
func (step labelJNZStep) String() string  { return fmt.Sprintf("jnz :%s", string(step)) }
func (step forkStep) String() string      { return fmt.Sprintf("fork @%d", int(step)) }
func (step relForkStep) String() string   { return fmt.Sprintf("fork %+d", int(step)) }
func (step labelForkStep) String() string { return fmt.Sprintf("fork :%s", string(step)) }

func (step labelJmpStep) annotate() string  { return fmt.Sprintf("-> :%s", string(step)) }
func (step labelJZStep) annotate() string   { return fmt.Sprintf("?-> :%s", string(step)) }
func (step labelJNZStep) annotate() string  { return fmt.Sprintf("?-> :%s", string(step)) }
func (step labelForkStep) annotate() string { return fmt.Sprintf("*-> :%s", string(step)) }

func (step jmpStep) run(sol *solution) {
	sol.stepi = int(step)
}
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

func (step relJMPStep) run(sol *solution) {
	sol.stepi += int(step)
}
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

func (step labelJmpStep) run(sol *solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}
func (step labelJZStep) run(sol *solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}
func (step labelJNZStep) run(sol *solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}

func (step forkStep) run(sol *solution) {
	child := sol.copy()
	child.stepi = int(step)
	sol.emit(child)
}
func (step relForkStep) run(sol *solution) {
	child := sol.copy()
	child.stepi += int(step)
	sol.emit(child)
}
func (step labelForkStep) run(sol *solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}

func (step labelJmpStep) resolveLabels(labels map[string]int) solutionStep {
	if addr, ok := labels[string(step)]; ok {
		return jmpStep(addr)
	}
	return nil
}
func (step labelJZStep) resolveLabels(labels map[string]int) solutionStep {
	if addr, ok := labels[string(step)]; ok {
		return jzStep(addr)
	}
	return nil
}
func (step labelJNZStep) resolveLabels(labels map[string]int) solutionStep {
	if addr, ok := labels[string(step)]; ok {
		return jnzStep(addr)
	}
	return nil
}
func (step labelForkStep) resolveLabels(labels map[string]int) solutionStep {
	if addr, ok := labels[string(step)]; ok {
		return forkStep(addr)
	}
	return nil
}

func isForkStep(step solutionStep) bool {
	switch step.(type) {
	case forkStep:
		return true
	case relForkStep:
		return true
	case labelForkStep:
		return true
	default:
		return false
	}
}

type finishStep string

func (step finishStep) String() string    { return fmt.Sprintf("HALT :%s", string(step)) }
func (step finishStep) run(sol *solution) { sol.exit(nil) }
func (step finishStep) labelName() string { return string(step) }
func (step finishStep) expandStep(
	addr int,
	parts [][]solutionStep,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]solutionStep, map[string]int) {
	if annotate != nil {
		annotate(addr,
			fmt.Sprintf(":%s", string(step)),
			"Normal Exit")
	}
	return addr + 1, append(parts, []solutionStep{exitStep{nil}}), labels
}

type rangeStep struct {
	label    string
	min, max int
}

func (step rangeStep) labelName() string {
	return step.label
}
func (step rangeStep) String() string {
	return fmt.Sprintf(":%s range [%d, %d]", step.label, step.min, step.max)
}
func (step rangeStep) run(sol *solution) {
	sol.exit(fmt.Errorf("unexpanded range :%s [%d, %d]", step.label, step.min, step.max))
}

func (step rangeStep) expandStep(
	addr int,
	parts [][]solutionStep,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]solutionStep, map[string]int) {
	if annotate != nil {
		bodySym := fmt.Sprintf("%s:body", step.label)
		nextSym := fmt.Sprintf("%s:next", step.label)
		contSym := fmt.Sprintf("%s:cont", step.label)
		annotate(addr, labelStep(step.label).String())
		annotate(addr, fmt.Sprintf("range:[%d, %d]", step.min, step.max))
		annotate(addr+2, labelStep(bodySym).String())
		annotate(addr+8, labelStep(nextSym).String())
		annotate(addr+17, labelStep(contSym).String())
		annotate(addr+4, labelJNZStep(nextSym).annotate())
		annotate(addr+5, labelForkStep(nextSym).annotate())
		annotate(addr+7, labelJmpStep(contSym).annotate())
		annotate(addr+12, labelJNZStep(bodySym).annotate())
		annotate(addr+15, labelJZStep(contSym).annotate())
	}
	return addr + 18, append(parts, []solutionStep{
		setAStep(step.min),       //  0: :LABEL ra = $min
		setCAStep{},              //  1: rc = ra
		setACStep{},              //  2: :LABEL:body ra = rc
		isUsedStep{},             //  3: used?
		jnzStep(addr + 8),        //  4: jnz :next
		forkStep(addr + 8),       //  5: fork :next
		setACStep{},              //  6: ra = rc
		jmpStep(addr + 17),       //  7: jmp :cont
		setACStep{},              //  8: :LABEL:next ra = rc
		addStep(1),               //  9: add 1
		setCAStep{},              // 10: rc = ra
		ltStep(step.max),         // 11: lt $max
		jnzStep(addr + 2),        // 12: jnz :body
		setACStep{},              // 13: ra = rc
		isUsedStep{},             // 14: used?
		jzStep(addr + 17),        // 15: jz :cont
		exitStep{errAlreadyUsed}, // 16: exit errAlreadyUsed
		setACStep{},              // 17: :LABEL:cont ra = rc
	}), labels
}
