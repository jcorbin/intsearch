package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	errAlreadyUsed = errors.New("value already used")
	errCheckFailed = errors.New("check failed")
	errNoChoices   = errors.New("no choices left")
)

type verifyError string

func isVerifyError(err error) bool {
	_, is := err.(verifyError)
	return is
}

func (ve verifyError) Error() string {
	return fmt.Sprintf("verify failed: %s", string(ve))
}

type goGen struct {
	*planProblem
	steps       []solutionStep
	carryPrior  *column
	carrySaved  bool
	carryValid  bool
	usedSymbols map[string]struct{}
	labels      map[string]int
	addrAnnos   map[int][]string
	lastLogDump int
}

func newGoGen(prob *planProblem) *goGen {
	gg := &goGen{
		planProblem: prob,
		usedSymbols: make(map[string]struct{}, 3*len(prob.letterSet)),
	}
	return gg
}

func fallFact(x, y int) int {
	z := 1
	for y > 0 {
		z *= x
		x--
		y--
	}
	return z
}

func (gg *goGen) searchInit(emit emitFunc) int {
	emit(newSolution(&gg.planProblem.problem, gg.steps, emit))
	// worst case, we have to run every step for every possible brute force solution
	numBrute := fallFact(gg.base, len(gg.letterSet))
	return numBrute * len(gg.steps)
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
	labels := gg.annosFor(i)
	if len(labels) == 0 {
		return ""
	}
	return strings.Join(labels, ", ")
}

func (gg *goGen) annosFor(addr int) []string {
	if gg.addrAnnos == nil {
		return nil
	}
	if addr > len(gg.steps) {
		return []string{"INVALID"}
	}
	if annos, ok := gg.addrAnnos[addr]; ok {
		return annos
	}
	return nil
}

func (gg *goGen) decorate(args []interface{}) []string {
	if gg.addrAnnos == nil {
		return nil
	}
	var dec []string
	for _, arg := range args {
		if sol, ok := arg.(*solution); ok {
			dec = append(dec, gg.annosFor(sol.stepi)...)
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

func (gg *goGen) stashCarry(col *column) {
	if gg.carryPrior == col && (col == nil || gg.carrySaved) {
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
		labelStep(gg.gensym("computeSum(%s)", col.label())))
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
		labelStep(gg.gensym("computeSummand(%s)", col.label())),
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
	switch col.carry {
	case carryZero:
		fallthrough
	case carryOne:
		if !gg.restoreCarry(col) {
			return
		}
	default:
		return
	}

	label := gg.gensym("checkFixedCarry(%s)", col.label())
	switch col.carry {
	case carryZero:
		gg.steps = append(gg.steps,
			labelStep(label),
			relJZStep(1),
			exitStep{errCheckFailed})
	case carryOne:
		gg.steps = append(gg.steps,
			labelStep(label),
			relJNZStep(1),
			exitStep{errCheckFailed})
	}
}

func (gg *goGen) chooseRange(c byte, min, max int) {
	gg.stashCarry(gg.carryPrior)
	gg.carryValid = false
	label := gg.gensym("choose(%s, %d, %d)", string(c), min, max)
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

	switch col.carry {
	case carryZero:
		fallthrough
	case carryOne:
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("ensureCarry(%d):fixed", col.i)),
			setAStep(col.carry))
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
		labelStep(gg.gensym("computeCarry(%d)", col.label())))
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

func (gg *goGen) check(err error) {
	gg.doVerify("check", err)
}

func (gg *goGen) checkColumn(col *column, err error) {
	if err == nil {
		err = errCheckFailed
	}

	a, b, c := col.cx[0], col.cx[1], col.cx[2]
	gg.ensureCarry(col.prior)
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("checkColumn(%s)", col.label())))
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
		exitStep{err})
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

func (gg *goGen) verify() {
	gg.doVerify("verify", nil)
}

func (gg *goGen) doVerify(name string, err error) {
	name = gg.gensym(name)
	gg.steps = append(gg.steps, labelStep(name))
	gg.verifyInitialLetters(name, err)
	gg.verifyDuplicateLetters(name, err)
	gg.verifyLettersNonNegative(name, err)
	gg.verifyColumns(name, err)
}

func (gg *goGen) verifyColumns(name string, err error) {
	// verify columns from bottom up
	for i := len(gg.columns) - 1; i >= 0; i-- {
		if gg.columns[i].unknown > 0 {
			return
		}
		col := &gg.columns[i]
		colErr := err
		if colErr == nil {
			colErr = verifyError(col.label())
		}
		gg.checkColumn(col, colErr)
	}

	// final carry may be constant by construction
	if step, ok := gg.steps[len(gg.steps)-1].(setAStep); ok {
		if int(step) != 0 {
			panic("broken final carry")
		}
		return
	}

	// final carry must be 0
	finErr := err
	if finErr == nil {
		finErr = verifyError("final carry must be 0")
	}
	gg.steps = append(gg.steps,
		relJZStep(1),
		exitStep{finErr})
}

func (gg *goGen) verifyInitialLetters(name string, err error) {
	if err == nil {
		err = verifyError("initial letter cannot be zero")
	}
	gg.steps = append(gg.steps, labelStep(gg.gensym("%s:initialLetters", name)))
	for _, word := range gg.words {
		gg.steps = append(gg.steps,
			loadStep(word[0]),
			relJNZStep(1),
			exitStep{err})
	}
}

func (gg *goGen) verifyDuplicateLetters(name string, err error) {
	if err == nil {
		err = verifyError("duplicate valued character")
	}
	gg.steps = append(gg.steps, labelStep(gg.gensym("%s:duplicateLetters", name)))
	letters := gg.sortedLetters()
	for i, c := range letters {
		if !gg.known[c] {
			continue
		}
		for j, d := range letters {
			if !gg.known[d] {
				continue
			}
			if j > i {
				gg.steps = append(gg.steps,
					loadStep(c),
					subValueStep(d),
					relJNZStep(1),
					exitStep{err})
			}
		}
	}
}

func (gg *goGen) verifyLettersNonNegative(name string, err error) {
	if err == nil {
		err = verifyError("negative valued character")
	}
	gg.steps = append(gg.steps, labelStep(gg.gensym("%s:allLettersNonNegative", name)))
	for _, c := range gg.sortedLetters() {
		if !gg.known[c] {
			continue
		}
		gg.steps = append(gg.steps,
			loadStep(c),
			ltStep(0),
			relJZStep(1),
			exitStep{err})
	}
}

func (gg *goGen) finish() {
	lastStep := gg.steps[len(gg.steps)-1]
	if _, isFinish := lastStep.(finishStep); isFinish {
		panic("double goGen.finish")
	}
	gg.steps = append(gg.steps, finishStep(gg.gensym("finish")))
}

func (gg *goGen) finalize() {
	gg.compile()
}

func (gg *goGen) takeAnnotation(addr int, annos ...string) {
	if gg.addrAnnos == nil {
		gg.addrAnnos = make(map[int][]string)
	}
	gg.addrAnnos[addr] = append(gg.addrAnnos[addr], annos...)
}

func (gg *goGen) compile() {
	var parts [][]solutionStep
	var addr int
	addr, parts, gg.labels = expandSteps(addr, gg.steps, nil, gg.labels, gg.takeAnnotation)
	steps := make([]solutionStep, 0, addr)
	for _, part := range parts {
		steps = append(steps, part...)
	}
	if len(steps) != addr {
		panic(fmt.Sprintf(
			"compiled final addr %d mismatches final step length %d by %d",
			addr, len(steps), addr-len(steps)))
	}
	gg.steps, gg.labels = resolveLabels(steps, gg.labels)
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

func printLastKSteps(k int, steps []solutionStep) {
	i := len(steps) - k - 1
	if i < 0 {
		i = 0
	}
	for j, step := range steps[i:] {
		fmt.Printf("[%v]: %v\n", i+j, step)
	}
}
