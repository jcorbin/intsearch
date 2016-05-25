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

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type ltAStep int
type ltBStep int
type ltCStep int
type lteAStep int
type lteBStep int
type lteCStep int
type eqAStep int
type eqBStep int
type eqCStep int
type gteAStep int
type gteBStep int
type gteCStep int
type gtAStep int
type gtBStep int
type gtCStep int

func (step ltAStep) String() string  { return fmt.Sprintf("lt ra, %v", int(step)) }
func (step ltBStep) String() string  { return fmt.Sprintf("lt rb, %v", int(step)) }
func (step ltCStep) String() string  { return fmt.Sprintf("lt rc, %v", int(step)) }
func (step lteAStep) String() string { return fmt.Sprintf("lte ra, %v", int(step)) }
func (step lteBStep) String() string { return fmt.Sprintf("lte rb, %v", int(step)) }
func (step lteCStep) String() string { return fmt.Sprintf("lte rc, %v", int(step)) }
func (step eqAStep) String() string  { return fmt.Sprintf("eq ra, %v", int(step)) }
func (step eqBStep) String() string  { return fmt.Sprintf("eq rb, %v", int(step)) }
func (step eqCStep) String() string  { return fmt.Sprintf("eq rc, %v", int(step)) }
func (step gteAStep) String() string { return fmt.Sprintf("gte ra, %v", int(step)) }
func (step gteBStep) String() string { return fmt.Sprintf("gte rb, %v", int(step)) }
func (step gteCStep) String() string { return fmt.Sprintf("gte rc, %v", int(step)) }
func (step gtAStep) String() string  { return fmt.Sprintf("gt ra, %v", int(step)) }
func (step gtBStep) String() string  { return fmt.Sprintf("gt rb, %v", int(step)) }
func (step gtCStep) String() string  { return fmt.Sprintf("gt rc, %v", int(step)) }

func (step ltAStep) run(sol *solution)  { sol.ra = boolInt(sol.ra < int(step)) }
func (step ltBStep) run(sol *solution)  { sol.ra = boolInt(sol.rb < int(step)) }
func (step ltCStep) run(sol *solution)  { sol.ra = boolInt(sol.rc < int(step)) }
func (step lteAStep) run(sol *solution) { sol.ra = boolInt(sol.ra <= int(step)) }
func (step lteBStep) run(sol *solution) { sol.ra = boolInt(sol.rb <= int(step)) }
func (step lteCStep) run(sol *solution) { sol.ra = boolInt(sol.rc <= int(step)) }
func (step eqAStep) run(sol *solution)  { sol.ra = boolInt(sol.ra == int(step)) }
func (step eqBStep) run(sol *solution)  { sol.ra = boolInt(sol.rb == int(step)) }
func (step eqCStep) run(sol *solution)  { sol.ra = boolInt(sol.rc == int(step)) }
func (step gteAStep) run(sol *solution) { sol.ra = boolInt(sol.ra >= int(step)) }
func (step gteBStep) run(sol *solution) { sol.ra = boolInt(sol.rb >= int(step)) }
func (step gteCStep) run(sol *solution) { sol.ra = boolInt(sol.rc >= int(step)) }
func (step gtAStep) run(sol *solution)  { sol.ra = boolInt(sol.ra > int(step)) }
func (step gtBStep) run(sol *solution)  { sol.ra = boolInt(sol.rb > int(step)) }
func (step gtCStep) run(sol *solution)  { sol.ra = boolInt(sol.rc > int(step)) }

type negAStep struct{}
type addARegBStep struct{}
type addARegCStep struct{}
type subARegBStep struct{}
type subARegCStep struct{}
type addAValueStep byte
type subAValueStep byte
type addAStep int
type subAStep int
type modAStep int
type divAStep int

func (step negAStep) String() string      { return fmt.Sprintf("negate ra") }
func (step addARegBStep) String() string  { return "add ra, rb" }
func (step addARegCStep) String() string  { return "add ra, rc" }
func (step subARegBStep) String() string  { return "sub ra, rb" }
func (step subARegCStep) String() string  { return "sub ra, rc" }
func (step addAValueStep) String() string { return fmt.Sprintf("add ra, $%s", string(step)) }
func (step subAValueStep) String() string { return fmt.Sprintf("sub ra, $%s", string(step)) }
func (step addAStep) String() string      { return fmt.Sprintf("add ra, %+d", int(step)) }
func (step subAStep) String() string      { return fmt.Sprintf("sub ra, %+d", int(step)) }
func (step modAStep) String() string      { return fmt.Sprintf("mod ra, %v", int(step)) }
func (step divAStep) String() string      { return fmt.Sprintf("div ra, %v", int(step)) }

func (step negAStep) run(sol *solution)      { sol.ra = -sol.ra }
func (step addARegBStep) run(sol *solution)  { sol.ra += sol.rb }
func (step addARegCStep) run(sol *solution)  { sol.ra += sol.rc }
func (step subARegBStep) run(sol *solution)  { sol.ra -= sol.rb }
func (step subARegCStep) run(sol *solution)  { sol.ra -= sol.rc }
func (step addAValueStep) run(sol *solution) { sol.ra += sol.values[step] }
func (step subAValueStep) run(sol *solution) { sol.ra -= sol.values[step] }
func (step addAStep) run(sol *solution)      { sol.ra += int(step) }
func (step subAStep) run(sol *solution)      { sol.ra -= int(step) }
func (step modAStep) run(sol *solution)      { sol.ra = (sol.ra + int(step)<<1) % int(step) }
func (step divAStep) run(sol *solution) {
	if sol.ra < 0 {
		sol.ra = -sol.ra / int(step)
	} else {
		sol.ra = sol.ra / int(step)
	}
}

type negBStep struct{}
type addBRegAStep struct{}
type addBRegCStep struct{}
type subBRegAStep struct{}
type subBRegCStep struct{}
type addBValueStep byte
type subBValueStep byte
type addBStep int
type subBStep int
type modBStep int
type divBStep int

func (step negBStep) String() string      { return fmt.Sprintf("negate rb") }
func (step addBRegAStep) String() string  { return "add rb, ra" }
func (step addBRegCStep) String() string  { return "add rb, rc" }
func (step subBRegAStep) String() string  { return "sub rb, ra" }
func (step subBRegCStep) String() string  { return "sub rb, rc" }
func (step addBValueStep) String() string { return fmt.Sprintf("add rb, $%s", string(step)) }
func (step subBValueStep) String() string { return fmt.Sprintf("sub rb, $%s", string(step)) }
func (step addBStep) String() string      { return fmt.Sprintf("add rb, %+d", int(step)) }
func (step subBStep) String() string      { return fmt.Sprintf("sub rb, %+d", int(step)) }
func (step modBStep) String() string      { return fmt.Sprintf("mod rb, %v", int(step)) }
func (step divBStep) String() string      { return fmt.Sprintf("div rb, %v", int(step)) }

func (step negBStep) run(sol *solution)      { sol.rb = -sol.rb }
func (step addBRegAStep) run(sol *solution)  { sol.rb += sol.ra }
func (step addBRegCStep) run(sol *solution)  { sol.rb += sol.rc }
func (step subBRegAStep) run(sol *solution)  { sol.rb -= sol.ra }
func (step subBRegCStep) run(sol *solution)  { sol.rb -= sol.rc }
func (step addBValueStep) run(sol *solution) { sol.rb += sol.values[step] }
func (step subBValueStep) run(sol *solution) { sol.rb -= sol.values[step] }
func (step addBStep) run(sol *solution)      { sol.rb += int(step) }
func (step subBStep) run(sol *solution)      { sol.rb -= int(step) }
func (step modBStep) run(sol *solution)      { sol.rb = (sol.rb + int(step)<<1) % int(step) }
func (step divBStep) run(sol *solution) {
	if sol.ra < 0 {
		sol.ra = -sol.ra / int(step)
	} else {
		sol.ra = sol.ra / int(step)
	}
}

type exitStep struct{ err error }

func (step exitStep) String() string    { return fmt.Sprintf("exit(%v)", step.err) }
func (step exitStep) run(sol *solution) { sol.exit(step.err) }

type usedAStep struct{}
type usedBStep struct{}
type usedCStep struct{}

func (step usedAStep) String() string { return fmt.Sprintf("used? ra") }
func (step usedBStep) String() string { return fmt.Sprintf("used? rb") }
func (step usedCStep) String() string { return fmt.Sprintf("used? rc") }

func (step usedAStep) run(sol *solution) {
	if sol.used[sol.ra] {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}
func (step usedBStep) run(sol *solution) {
	if sol.used[sol.rb] {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}
func (step usedCStep) run(sol *solution) {
	if sol.used[sol.rc] {
		sol.ra = 1
	} else {
		sol.ra = 0
	}
}

type storeAStep byte
type storeBStep byte
type storeCStep byte
type loadAStep byte
type loadBStep byte
type loadCStep byte

func (c storeAStep) String() string { return fmt.Sprintf("store %s, ra", string(c)) }
func (c storeBStep) String() string { return fmt.Sprintf("store %s, rb", string(c)) }
func (c storeCStep) String() string { return fmt.Sprintf("store %s, rc", string(c)) }
func (c loadAStep) String() string  { return fmt.Sprintf("load ra, %s", string(c)) }
func (c loadBStep) String() string  { return fmt.Sprintf("load rb, %s", string(c)) }
func (c loadCStep) String() string  { return fmt.Sprintf("load rc, %s", string(c)) }
func (c storeAStep) run(sol *solution) {
	sol.values[c] = sol.ra
	sol.used[sol.ra] = true
}
func (c storeBStep) run(sol *solution) {
	sol.values[c] = sol.rb
	sol.used[sol.rb] = true
}
func (c storeCStep) run(sol *solution) {
	sol.values[c] = sol.rc
	sol.used[sol.rc] = true
}
func (c loadAStep) run(sol *solution) {
	sol.ra = sol.values[c]
}
func (c loadBStep) run(sol *solution) {
	sol.rb = sol.values[c]
}
func (c loadCStep) run(sol *solution) {
	sol.rc = sol.values[c]
}

func isStoreStep(step solutionStep) bool {
	switch step.(type) {
	case storeAStep:
		return true
	case storeBStep:
		return true
	case storeCStep:
		return true
	default:
		return false
	}
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
		annotate(addr, fmt.Sprintf(":%s", step.label))
		annotate(addr, fmt.Sprintf("range:[%d, %d]", step.min, step.max))
		annotate(addr+2, fmt.Sprintf(":%s:body", step.label))
		annotate(addr+8, fmt.Sprintf(":%s:next", step.label))
		annotate(addr+17, fmt.Sprintf(":%s:cont", step.label))
	}
	return addr + 18, append(parts, []solutionStep{
		setAStep(step.min),       //  0: :LABEL ra = $min
		setBAStep{},              //  1: rb = ra
		setABStep{},              //  2: :LABEL:body ra = rb
		usedAStep{},              //  3: used?
		jnzStep(addr + 8),        //  4: jnz :next
		forkStep(addr + 8),       //  5: fork :next
		setABStep{},              //  6: ra = rb
		jmpStep(addr + 17),       //  7: jmp :cont
		setABStep{},              //  8: :LABEL:next ra = rb
		addAStep(1),              //  9: add 1
		setBAStep{},              // 10: rb = ra
		ltAStep(step.max),        // 11: lt $max
		jnzStep(addr + 2),        // 12: jnz :body
		setABStep{},              // 13: ra = rb
		usedAStep{},              // 14: used?
		jzStep(addr + 17),        // 15: jz :cont
		exitStep{errAlreadyUsed}, // 16: exit errAlreadyUsed
		setABStep{},              // 17: :LABEL:cont ra = rb
	}), labels
}
