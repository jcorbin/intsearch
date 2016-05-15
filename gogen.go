package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	errAlreadyUsed    = errors.New("value already used")
	errCheckFailed    = errors.New("check failed")
	errNegativeValue  = errors.New("negative valued character")
	errDuplicateValue = errors.New("duplicate valued character")
	errVerifyFailed   = errors.New("verify failed")
	errNoChoices      = errors.New("no choices left")
)

type goGen struct {
	*planProblem
	steps        []solutionStep
	verified     bool
	useForkUntil bool
	carryPrior   *column
	carryFixed   map[int]int
	carrySaved   bool
	carryValid   bool
	usedSymbols  map[string]struct{}
	labels       map[string]int
	addrLabels   [][]string
	lastLogDump  int
}

func newGoGen(prob *planProblem, verified bool) *goGen {
	gg := &goGen{
		planProblem: prob,
		carryFixed:  make(map[int]int, prob.numColumns()),
		usedSymbols: make(map[string]struct{}, 3*len(prob.letterSet)),
		verified:    verified,
	}
	return gg
}

func (gg *goGen) searchInit(emit emitFunc) {
	emit(newSolution(&gg.planProblem.problem, gg.steps, emit))
}

func (gg *goGen) loggedGen() solutionGen {
	return multiGen([]solutionGen{
		newLogGen(gg.planProblem),
		gg,
		// TODO: doesn't work right wrt fork alt
		// afterGen(gg.dumpLastSteps),
	})
}

func (gg *goGen) dumpLastSteps() {
	// TODO: don't dump after finalize when we resurrect this
	i := gg.lastLogDump
	for ; i < len(gg.steps); i++ {
		fmt.Printf("%v: %v\n", i, gg.steps[i])
	}
	if i > gg.lastLogDump {
		fmt.Println()
		gg.lastLogDump = i
	}
}

func (gg *goGen) labelFor(i int) string {
	labels := gg.labelsFor(i)
	if len(labels) == 0 {
		return ""
	}
	return strings.Join(labels, ", ")
}

func (gg *goGen) labelsFor(i int) []string {
	if gg.addrLabels == nil {
		return nil
	}
	if i < len(gg.addrLabels) {
		return gg.addrLabels[i]
	}
	return []string{"INVALID"}
}

func (gg *goGen) decorate(args []interface{}) []string {
	var dec []string
	if gg.addrLabels != nil {
		for _, arg := range args {
			if sol, ok := arg.(*solution); ok {
				dec = append(dec, gg.labelsFor(sol.stepi)...)
			}
		}
	}
	return dec
}

func (gg *goGen) logf(format string, args ...interface{}) error {
	return nil
}

func (gg *goGen) init(desc string) {
}

func (gg *goGen) fix(c byte, v int) {
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("fix(%s)", string(c))),
		setAStep(v),
		storeStep(c))
}

func (gg *goGen) fixCarry(i, v int) {
	gg.carryFixed[i] = v
}

func (gg *goGen) stashCarry(col *column) {
	if gg.carryPrior == col && gg.carrySaved {
		return
	}

	if col == nil {
		gg.carrySaved = false
		gg.carryPrior = nil
		return
	}

	if !gg.carryValid {
		return
	}

	gg.steps = append(gg.steps,
		labelStep(gg.gensym("stashCarry(%d)", col.i)),
		setBAStep{})
	gg.carrySaved = true
	gg.carryPrior = col
}

func (gg *goGen) saveCarry(col *column) {
	if !gg.carryValid {
		gg.ensureCarry(col)
	}
	gg.stashCarry(col)
}

func (gg *goGen) computeSum(col *column) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for c:
	//   c = carry + a + b (mod base)
	a, b, c := col.cx[0], col.cx[1], col.cx[2]
	gg.ensureCarry(col.prior)
	gg.carryValid = false
	gg.carrySaved = false

	steps := make([]solutionStep, 0, 8)
	steps = append(steps,
		labelStep(gg.gensym("computeSum(%s)", charsLabel(a, b, c))))
	if a != 0 {
		steps = append(steps, addValueStep(a))
	}
	if b != 0 {
		steps = append(steps, addValueStep(b))
	}
	steps = append(steps,
		setBAStep{},
		modStep(gg.base),
		storeStep(c),
		setABStep{},
		divStep(gg.base))
	gg.steps = append(gg.steps, steps...)

	gg.carryPrior = col
	gg.carryValid = true
	gg.carrySaved = false

	gg.checkAfterCompute(col, c)
}

func (gg *goGen) computeFirstSummand(col *column) {
	gg.computeSummand(col, col.cx[0], col.cx[1], col.cx[2])
}

func (gg *goGen) computeSecondSummand(col *column) {
	gg.computeSummand(col, col.cx[1], col.cx[0], col.cx[2])
}

func (gg *goGen) computeSummand(col *column, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for a:
	//   a = c - b - carry (mod base)
	gg.saveCarry(col.prior)
	gg.carryValid = false

	steps := make([]solutionStep, 0, 9)
	steps = append(steps,
		labelStep(gg.gensym("computeSummand(%s)", charsLabel(a, b, c))),
		negateStep{})
	if c != 0 {
		steps = append(steps, addValueStep(c))
	}
	if b != 0 {
		steps = append(steps, subValueStep(b))
	}
	steps = append(steps,
		modStep(gg.base),
		storeStep(a),
		addRegBStep{})
	if b != 0 {
		steps = append(steps, addValueStep(b))
	}
	steps = append(steps, divStep(gg.base))
	gg.steps = append(gg.steps, steps...)

	gg.carryPrior = col
	gg.carryValid = true
	gg.carrySaved = false

	gg.checkAfterCompute(col, a)
}

func (gg *goGen) checkAfterCompute(col *column, c byte) {
	if c == gg.words[0][0] || c == gg.words[1][0] || c == gg.words[2][0] {
		gg.checkInitialLetter(col, c)
	}
	gg.checkFixedCarry(col)
}

func (gg *goGen) checkInitialLetter(col *column, c byte) {
	steps := make([]solutionStep, 0, 4)
	steps = append(steps,
		labelStep(gg.gensym("checkInitialLetter(%s)", string(c))))
	if gg.carryValid {
		gg.stashCarry(col)
		gg.carryValid = false
		steps = append(steps, loadStep(c))
	}
	steps = append(steps,
		relJNZStep(1),
		exitStep{errCheckFailed})
	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) checkFixedCarry(col *column) {
	if col.i <= 0 {
		return
	}
	carryOut, fixed := gg.carryFixed[col.i-1]
	if !fixed {
		return
	}
	if !gg.restoreCarry(col) {
		return
	}

	label := gg.gensym("checkFixedCarry(%s)", charsLabel(col.cx[0], col.cx[1], col.cx[2]))
	if carryOut == 0 {
		gg.steps = append(gg.steps,
			labelStep(label),
			relJZStep(1),
			exitStep{errCheckFailed})
	} else {
		gg.steps = append(gg.steps,
			labelStep(label),
			relJNZStep(1),
			exitStep{errCheckFailed})
	}
}

func (gg *goGen) chooseRange(col *column, c byte, i, min, max int) {
	gg.stashCarry(col.prior)
	gg.carryValid = false

	label := gg.gensym("choose(%s, %d, %d)", string(c), min, max)
	if gg.useForkUntil {
		gg.steps = append(gg.steps,
			labelStep(label),   // :choose($c)
			setAStep(min),      // ra = $min
			forkUntilStep(max), // forUntil $max
			storeStep(c),       // store $c
		)
	} else {
		var (
			loopSym     = gg.gensym("choose(%s):loop", string(c))
			nextLoopSym = gg.gensym("choose(%s):nextLoop", string(c))
			contSym     = gg.gensym("choose(%s):cont", string(c))
		)
		gg.steps = append(gg.steps,
			labelStep(label),           // :choose($c)
			setAStep(min),              // ra = $min
			setCAStep{},                // rc = ra
			labelStep(loopSym),         // :loop
			setACStep{},                // ra = rc
			isUsedStep{},               // used?
			labelJNZStep(nextLoopSym),  // jnz :next_loop
			forkLabelStep(nextLoopSym), // fork :next_loop
			setACStep{},                // ra = rc
			labelJmpStep(contSym),      // jmp :cont
			labelStep(nextLoopSym),     // :nextLoop
			setACStep{},                // ra = rc
			addStep(1),                 // add 1
			setCAStep{},                // rc = ra
			ltStep(max),                // lt $max
			labelJNZStep(loopSym),      // jnz :loop
			setACStep{},                // ra = rc
			isUsedStep{},               // used?
			labelJZStep(contSym),       // jz :cont
			exitStep{errAlreadyUsed},   // exit errAlreadyUsed
			labelStep(contSym),         // :cont
			setACStep{},                // ra = rc
			storeStep(c),               // store $c
		)
	}
}

func (gg *goGen) restoreCarry(col *column) bool {
	if col != gg.carryPrior {
		return false
	}
	if gg.carryValid {
		return true
	}
	if !gg.carrySaved {
		return false
	}
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("restoreCarry(%d)", col.i)),
		setABStep{})
	gg.carryValid = true
	return true
}

func (gg *goGen) ensureCarry(col *column) {
	if col == nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("ensureCarry:nil")),
			setAStep(0))
		gg.carryPrior = nil
		gg.carrySaved = false
		gg.carryValid = true
		return
	}

	if value, ok := gg.carryFixed[col.i]; ok {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("ensureCarry(%d):fixed", col.i)),
			setAStep(value))
		gg.carryPrior = col
		gg.carrySaved = false
		gg.carryValid = true
		return
	}

	if gg.restoreCarry(col) {
		return
	}

	c1 := col.cx[0]
	if c1 != 0 && !gg.known[c1] {
		log.Fatalf("cannot compute carry from unknown c1 for column %v", col)
	}

	c2 := col.cx[1]
	if c2 != 0 && !gg.known[c2] {
		log.Fatalf("cannot compute carry from unknown c2 for column %v", col)
	}

	gg.ensureCarry(col.prior)
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("ensureCarry(%d):compute(%s)", col.i, charsLabel(c1, c2))))
	steps := make([]solutionStep, 0, 3)
	if c1 != 0 {
		steps = append(steps, addValueStep(c1))
	}
	if c2 != 0 {
		steps = append(steps, addValueStep(c2))
	}
	steps = append(steps, divStep(gg.base))
	gg.steps = append(gg.steps, steps...)

	gg.carryPrior = col
	gg.carrySaved = false
	gg.carryValid = true
}

func (gg *goGen) checkColumn(col *column) {
	a, b, c := col.cx[0], col.cx[1], col.cx[2]
	gg.ensureCarry(col.prior)
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("checkColumn(%s)", charsLabel(a, b, c))))
	steps := make([]solutionStep, 0, 9)

	n := 0
	if a != 0 {
		n++
		steps = append(steps, addValueStep(a))
	}
	if b != 0 {
		n++
		steps = append(steps, addValueStep(b))
	}
	if n > 0 {
		steps = append(steps,
			setCAStep{},
			modStep(gg.base))
	}
	steps = append(steps,
		subValueStep(c),
		relJZStep(1),
		exitStep{errCheckFailed})
	if n > 0 {
		steps = append(steps,
			setACStep{},
			divStep(gg.base))
	} else {
		steps = append(steps, setAStep(0))
	}
	gg.carryPrior = col
	gg.carrySaved = false
	gg.carryValid = true
	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) verifySteps() []solutionStep {
	N := len(gg.letterSet)
	C := gg.numColumns()
	steps := make([]solutionStep, 0, N*N/2*4+N*4+1+C*9+2)

	letters := make([]byte, 0, N)
	for c := range gg.letterSet {
		letters = append(letters, c)
	}

	for i, c := range letters {
		for j, d := range letters {
			if j > i {
				steps = append(steps,
					loadStep(c),
					subValueStep(d),
					relJNZStep(1),
					exitStep{errDuplicateValue})
			}
		}
	}

	for _, c := range letters {
		steps = append(steps,
			loadStep(c),
			ltStep(0),
			relJZStep(1),
			exitStep{errNegativeValue})
	}

	steps = append(steps, setAStep(0))
	for i := C - 1; i >= 0; i-- {
		cx := gg.getColumn(i)
		if cx[0] != 0 {
			steps = append(steps, addValueStep(cx[0]))
		}
		if cx[1] != 0 {
			steps = append(steps, addValueStep(cx[1]))
		}
		steps = append(steps,
			setBAStep{},
			modStep(gg.base),
			subValueStep(cx[2]),
			relJZStep(1),
			exitStep{errVerifyFailed},
			setABStep{},
			divStep(gg.base))
	}
	steps = append(steps,
		relJZStep(1),
		exitStep{errVerifyFailed})

	return steps
}

func (gg *goGen) finish() {
	lastStep := gg.steps[len(gg.steps)-1]

	if gg.verified {
		if _, isVerifyJmp := lastStep.(labelJmpStep); isVerifyJmp {
			panic("double goGen.finish")
		}
		gg.steps = append(gg.steps, labelJmpStep("verify"))
		return
	}

	if _, isFinish := lastStep.(finishStep); isFinish {
		panic("double goGen.finish")
	}
	gg.steps = append(gg.steps, finishStep(gg.gensym("finish")))
}

func (gg *goGen) finalize() {
	if gg.verified {
		gg.steps = append(gg.steps, labelStep("verify"))
		gg.steps = append(gg.steps, gg.verifySteps()...)
		gg.steps = append(gg.steps, finishStep(gg.gensym("finish")))
	}

	gg.compile()
}

func (gg *goGen) compile() {
	var parts [][]solutionStep
	var addr int
	addr, parts, gg.labels = eraseLabels(addr, gg.steps, nil, gg.labels)
	steps := make([]solutionStep, 0, addr)
	for _, part := range parts {
		steps = append(steps, part...)
	}
	gg.steps, gg.labels = resolveLabels(steps, gg.labels)

	gg.addrLabels = make([][]string, len(gg.steps))
	for label, addr := range gg.labels {
		gg.addrLabels[addr] = append(gg.addrLabels[addr], label)
	}
}

func (gg *goGen) gensym(format string, args ...interface{}) string {
	name := fmt.Sprintf(format, args...)

	if _, used := gg.usedSymbols[name]; !used {
		gg.usedSymbols[name] = struct{}{}
		return name
	}

	i := 2
	for {
		sym := fmt.Sprintf("%s_%d", name, i)
		if _, used := gg.usedSymbols[sym]; !used {
			gg.usedSymbols[sym] = struct{}{}
			return sym
		}
		i++
	}
}

func charsLabel(cs ...byte) string {
	if len(cs) == 1 {
		return charLabel(cs[0])
	}
	ss := make([]string, len(cs))
	for i, c := range cs {
		ss[i] = charsLabel(c)
	}
	return strings.Join(ss, ", ")
}

func charLabel(c byte) string {
	if c == 0 {
		return "nil"
	}
	return string(c)
}

func printLastKSteps(k int, steps []solutionStep) {
	i := len(steps) - k - 1
	if i < 0 {
		i = 0
	}
	for j, step := range steps[i:] {
		fmt.Printf("[%v]: %v\n", i+j, step)
	}
}
