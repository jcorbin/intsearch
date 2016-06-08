package runnable

import (
	"errors"
	"fmt"
)

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

func (step setABStep) run(sol *Solution) { sol.ra = sol.rb }
func (step setACStep) run(sol *Solution) { sol.ra = sol.rc }
func (step setBAStep) run(sol *Solution) { sol.rb = sol.ra }
func (step setBCStep) run(sol *Solution) { sol.rb = sol.rc }
func (step setCAStep) run(sol *Solution) { sol.rc = sol.ra }
func (step setCBStep) run(sol *Solution) { sol.rc = sol.rb }

type setAStep int
type setBStep int
type setCStep int

func (step setAStep) String() string    { return fmt.Sprintf("ra = %v", int(step)) }
func (step setBStep) String() string    { return fmt.Sprintf("rb = %v", int(step)) }
func (step setCStep) String() string    { return fmt.Sprintf("rc = %v", int(step)) }
func (step setAStep) run(sol *Solution) { sol.ra = int(step) }
func (step setBStep) run(sol *Solution) { sol.rb = int(step) }
func (step setCStep) run(sol *Solution) { sol.rc = int(step) }

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

func (step ltAStep) run(sol *Solution)  { sol.ra = boolInt(sol.ra < int(step)) }
func (step ltBStep) run(sol *Solution)  { sol.ra = boolInt(sol.rb < int(step)) }
func (step ltCStep) run(sol *Solution)  { sol.ra = boolInt(sol.rc < int(step)) }
func (step lteAStep) run(sol *Solution) { sol.ra = boolInt(sol.ra <= int(step)) }
func (step lteBStep) run(sol *Solution) { sol.ra = boolInt(sol.rb <= int(step)) }
func (step lteCStep) run(sol *Solution) { sol.ra = boolInt(sol.rc <= int(step)) }
func (step eqAStep) run(sol *Solution)  { sol.ra = boolInt(sol.ra == int(step)) }
func (step eqBStep) run(sol *Solution)  { sol.ra = boolInt(sol.rb == int(step)) }
func (step eqCStep) run(sol *Solution)  { sol.ra = boolInt(sol.rc == int(step)) }
func (step gteAStep) run(sol *Solution) { sol.ra = boolInt(sol.ra >= int(step)) }
func (step gteBStep) run(sol *Solution) { sol.ra = boolInt(sol.rb >= int(step)) }
func (step gteCStep) run(sol *Solution) { sol.ra = boolInt(sol.rc >= int(step)) }
func (step gtAStep) run(sol *Solution)  { sol.ra = boolInt(sol.ra > int(step)) }
func (step gtBStep) run(sol *Solution)  { sol.ra = boolInt(sol.rb > int(step)) }
func (step gtCStep) run(sol *Solution)  { sol.ra = boolInt(sol.rc > int(step)) }

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

func (step negAStep) run(sol *Solution)      { sol.ra = -sol.ra }
func (step addARegBStep) run(sol *Solution)  { sol.ra += sol.rb }
func (step addARegCStep) run(sol *Solution)  { sol.ra += sol.rc }
func (step subARegBStep) run(sol *Solution)  { sol.ra -= sol.rb }
func (step subARegCStep) run(sol *Solution)  { sol.ra -= sol.rc }
func (step addAValueStep) run(sol *Solution) { sol.ra += sol.values[step] }
func (step subAValueStep) run(sol *Solution) { sol.ra -= sol.values[step] }
func (step addAStep) run(sol *Solution)      { sol.ra += int(step) }
func (step subAStep) run(sol *Solution)      { sol.ra -= int(step) }
func (step modAStep) run(sol *Solution)      { sol.ra = (sol.ra + int(step)<<1) % int(step) }
func (step divAStep) run(sol *Solution) {
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

func (step negBStep) run(sol *Solution)      { sol.rb = -sol.rb }
func (step addBRegAStep) run(sol *Solution)  { sol.rb += sol.ra }
func (step addBRegCStep) run(sol *Solution)  { sol.rb += sol.rc }
func (step subBRegAStep) run(sol *Solution)  { sol.rb -= sol.ra }
func (step subBRegCStep) run(sol *Solution)  { sol.rb -= sol.rc }
func (step addBValueStep) run(sol *Solution) { sol.rb += sol.values[step] }
func (step subBValueStep) run(sol *Solution) { sol.rb -= sol.values[step] }
func (step addBStep) run(sol *Solution)      { sol.rb += int(step) }
func (step subBStep) run(sol *Solution)      { sol.rb -= int(step) }
func (step modBStep) run(sol *Solution)      { sol.rb = (sol.rb + int(step)<<1) % int(step) }
func (step divBStep) run(sol *Solution) {
	if sol.ra < 0 {
		sol.ra = -sol.ra / int(step)
	} else {
		sol.ra = sol.ra / int(step)
	}
}

type exitStep struct{ err error }

func (step exitStep) String() string    { return fmt.Sprintf("exit(%v)", step.err) }
func (step exitStep) run(sol *Solution) { sol.exit(step.err) }

type usedAStep struct{}
type usedBStep struct{}
type usedCStep struct{}

func (step usedAStep) String() string { return fmt.Sprintf("used? ra") }
func (step usedBStep) String() string { return fmt.Sprintf("used? rb") }
func (step usedCStep) String() string { return fmt.Sprintf("used? rc") }

func (step usedAStep) run(sol *Solution) { sol.ra = boolInt(sol.used[sol.ra]) }
func (step usedBStep) run(sol *Solution) { sol.ra = boolInt(sol.used[sol.rb]) }
func (step usedCStep) run(sol *Solution) { sol.ra = boolInt(sol.used[sol.rc]) }

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
func (c storeAStep) run(sol *Solution) {
	sol.values[c] = sol.ra
	sol.used[sol.ra] = true
}
func (c storeBStep) run(sol *Solution) {
	sol.values[c] = sol.rb
	sol.used[sol.rb] = true
}
func (c storeCStep) run(sol *Solution) {
	sol.values[c] = sol.rc
	sol.used[sol.rc] = true
}
func (c loadAStep) run(sol *Solution) {
	sol.ra = sol.values[c]
}
func (c loadBStep) run(sol *Solution) {
	sol.rb = sol.values[c]
}
func (c loadCStep) run(sol *Solution) {
	sol.rc = sol.values[c]
}

func isStoreStep(step Step) bool {
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
type branchStep int
type relBranchStep int
type labelBranchStep string
type relFZStep int
type relFNZStep int
type relBZStep int
type relBNZStep int

func (step jmpStep) String() string         { return fmt.Sprintf("jmp @%d", int(step)) }
func (step jzStep) String() string          { return fmt.Sprintf("jz @%d", int(step)) }
func (step jnzStep) String() string         { return fmt.Sprintf("jnz @%d", int(step)) }
func (step relJMPStep) String() string      { return fmt.Sprintf("jmp %+d", int(step)) }
func (step relJZStep) String() string       { return fmt.Sprintf("jz %+d", int(step)) }
func (step relJNZStep) String() string      { return fmt.Sprintf("jnz %+d", int(step)) }
func (step labelJmpStep) String() string    { return fmt.Sprintf("jmp :%s", string(step)) }
func (step labelJZStep) String() string     { return fmt.Sprintf("jz :%s", string(step)) }
func (step labelJNZStep) String() string    { return fmt.Sprintf("jnz :%s", string(step)) }
func (step forkStep) String() string        { return fmt.Sprintf("fork @%d", int(step)) }
func (step relForkStep) String() string     { return fmt.Sprintf("fork %+d", int(step)) }
func (step labelForkStep) String() string   { return fmt.Sprintf("fork :%s", string(step)) }
func (step branchStep) String() string      { return fmt.Sprintf("branch @%d", int(step)) }
func (step relBranchStep) String() string   { return fmt.Sprintf("branch %+d", int(step)) }
func (step labelBranchStep) String() string { return fmt.Sprintf("branch :%s", string(step)) }
func (step relFZStep) String() string       { return fmt.Sprintf("fz %+d", int(step)) }
func (step relFNZStep) String() string      { return fmt.Sprintf("fnz %+d", int(step)) }
func (step relBZStep) String() string       { return fmt.Sprintf("bz %+d", int(step)) }
func (step relBNZStep) String() string      { return fmt.Sprintf("bnz %+d", int(step)) }

func (step labelJmpStep) annotate() string    { return fmt.Sprintf("-> :%s", string(step)) }
func (step labelJZStep) annotate() string     { return fmt.Sprintf("?-> :%s", string(step)) }
func (step labelJNZStep) annotate() string    { return fmt.Sprintf("?-> :%s", string(step)) }
func (step labelForkStep) annotate() string   { return fmt.Sprintf("*-> :%s", string(step)) }
func (step labelBranchStep) annotate() string { return fmt.Sprintf("/-> :%s", string(step)) }

func (step jmpStep) run(sol *Solution) {
	sol.stepi = int(step)
}
func (step jzStep) run(sol *Solution) {
	if sol.ra == 0 {
		sol.stepi = int(step)
	}
}
func (step jnzStep) run(sol *Solution) {
	if sol.ra != 0 {
		sol.stepi = int(step)
	}
}

func (step relJMPStep) run(sol *Solution) {
	sol.stepi += int(step)
}
func (step relJZStep) run(sol *Solution) {
	if sol.ra == 0 {
		sol.stepi += int(step)
	}
}
func (step relJNZStep) run(sol *Solution) {
	if sol.ra != 0 {
		sol.stepi += int(step)
	}
}

func (step labelJmpStep) run(sol *Solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}
func (step labelJZStep) run(sol *Solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}
func (step labelJNZStep) run(sol *Solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}

func (step forkStep) run(sol *Solution) {
	child := sol.copy()
	child.stepi = int(step)
	sol.emit(child)
}
func (step relForkStep) run(sol *Solution) {
	child := sol.copy()
	child.stepi += int(step)
	sol.emit(child)
}
func (step labelForkStep) run(sol *Solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}
func (step branchStep) run(sol *Solution) {
	child := sol.copy()
	sol.emit(child)
	sol.stepi = int(step)
}
func (step relBranchStep) run(sol *Solution) {
	child := sol.copy()
	sol.emit(child)
	sol.stepi += int(step)
}
func (step labelBranchStep) run(sol *Solution) {
	sol.exit(fmt.Errorf("unresolved label jump :%s", string(step)))
}

func (step relFZStep) run(sol *Solution) {
	if sol.ra == 0 {
		child := sol.copy()
		child.stepi += int(step)
		sol.emit(child)
	}
}
func (step relFNZStep) run(sol *Solution) {
	if sol.ra != 0 {
		child := sol.copy()
		child.stepi += int(step)
		sol.emit(child)
	}
}

func (step relBZStep) run(sol *Solution) {
	if sol.ra == 0 {
		child := sol.copy()
		sol.emit(child)
		sol.stepi += int(step)
	}
}
func (step relBNZStep) run(sol *Solution) {
	if sol.ra != 0 {
		child := sol.copy()
		sol.emit(child)
		sol.stepi += int(step)
	}
}

func (step labelJmpStep) resolveLabels(labels map[string]int) Step {
	if addr, ok := labels[string(step)]; ok {
		return jmpStep(addr)
	}
	return nil
}
func (step labelJZStep) resolveLabels(labels map[string]int) Step {
	if addr, ok := labels[string(step)]; ok {
		return jzStep(addr)
	}
	return nil
}
func (step labelJNZStep) resolveLabels(labels map[string]int) Step {
	if addr, ok := labels[string(step)]; ok {
		return jnzStep(addr)
	}
	return nil
}
func (step labelForkStep) resolveLabels(labels map[string]int) Step {
	if addr, ok := labels[string(step)]; ok {
		return forkStep(addr)
	}
	return nil
}
func (step labelBranchStep) resolveLabels(labels map[string]int) Step {
	if addr, ok := labels[string(step)]; ok {
		return branchStep(addr)
	}
	return nil
}

func isForkStep(step Step) bool {
	switch step.(type) {
	case forkStep:
		return true
	case relForkStep:
		return true
	case labelForkStep:
		return true
	case branchStep:
		return true
	case relBranchStep:
		return true
	case labelBranchStep:
		return true
	case relFZStep:
		return true
	case relFNZStep:
		return true
	case relBZStep:
		return true
	case relBNZStep:
		return true
	default:
		return false
	}
}

var errDeadFork = errors.New("dead fork")

type forkAltStep struct {
	alt       *StepGen
	name      string
	altLabel  string
	contLabel string
}

func (step forkAltStep) labelName() string { return step.name }
func (step forkAltStep) String() string {
	if step.name == "" {
		return "forkAlt UNNAMED"
	}
	return fmt.Sprintf("forkAlt :%s", step.name)
}
func (step forkAltStep) run(sol *Solution) {
	panic(fmt.Sprintf("unexpanded forkAlt :%s", step.name))
}
func (step forkAltStep) expandStep(
	addr int,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]Step, map[string]int) {
	// expands to:
	// fork :$name:cont
	// :$name:alt
	// ...
	// ... alt.steps
	// ...
	// exit errDeadFork
	// :$name:cont
	if step.alt.labels != nil {
		panic("double alt expand")
	}
	step.alt.labels = labels

	if annotate != nil {
		annotate(addr, labelForkStep(step.contLabel).annotate())
	}

	forkPart := []Step{
		labelForkStep(step.contLabel)}
	addr, parts = addr+1, append(parts, forkPart)

	if step.altLabel != "" {
		labels[step.altLabel] = addr
		if annotate != nil {
			annotate(addr, fmt.Sprintf(":%s", step.altLabel))
		}
	}

	altAddr := addr
	addr, parts, labels = expandSteps(addr, step.alt.steps, parts, labels, annotate)
	addr, parts = addr+1, append(parts, []Step{
		exitStep{errDeadFork}})

	forkPart[0] = relForkStep(addr - altAddr)

	if step.contLabel != "" {
		labels[step.contLabel] = addr
		if annotate != nil {
			annotate(addr, fmt.Sprintf(":%s", step.contLabel))
		}
	}

	return addr, parts, labels
}

type finishStep string

func (step finishStep) String() string    { return fmt.Sprintf("HALT :%s", string(step)) }
func (step finishStep) run(sol *Solution) { sol.exit(nil) }
func (step finishStep) labelName() string { return string(step) }
func (step finishStep) expandStep(
	addr int,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]Step, map[string]int) {
	if annotate != nil {
		annotate(addr,
			fmt.Sprintf(":%s", string(step)),
			"Normal Exit")
	}
	return addr + 1, append(parts, []Step{exitStep{nil}}), labels
}

type loopBStep struct {
	offset int
	max    int
}

func (step loopBStep) String() string {
	return fmt.Sprintf("loop %+d rb < %v", step.offset, step.max)
}
func (step loopBStep) run(sol *Solution) {
	sol.rb++
	if sol.rb < step.max {
		sol.stepi += step.offset
	}
}

type rangeStep struct {
	label    string
	min, max int
}

func (step rangeStep) labelName() string {
	return step.label
}
func (step rangeStep) String() string {
	if step.label == "" {
		return fmt.Sprintf("range [%d, %d]", step.min, step.max)
	}
	return fmt.Sprintf(":%s range [%d, %d]", step.label, step.min, step.max)
}
func (step rangeStep) run(sol *Solution) {
	sol.exit(fmt.Errorf("unexpanded %v", step))
}

func (step rangeStep) expandStep(
	addr int,
	parts [][]Step,
	labels map[string]int,
	annotate annoFunc,
) (int, [][]Step, map[string]int) {
	if annotate != nil && step.label != "" {
		annotate(addr,
			fmt.Sprintf(":%s", step.label),
			fmt.Sprintf("range:[%d, %d]", step.min, step.max))
		annotate(addr+1, fmt.Sprintf(":%s:body", step.label))
		annotate(addr+7, fmt.Sprintf(":%s:cont", step.label))
	}
	return addr + 7, append(parts, []Step{
		setBStep(step.min),       // 0: :LABEL rb = $min
		usedBStep{},              // 1: :LABEL:body used? rb
		relBZStep(4),             // 2: bz +4
		loopBStep{-3, step.max},  // 3: loop :body rb < $max
		usedBStep{},              // 4: used? rb
		relJZStep(1),             // 5: jz :cont
		exitStep{errAlreadyUsed}, // 6: exit errAlreadyUsed
		//                           7: :LABEL:cont
	}), labels
}
