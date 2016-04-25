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
	carrySaved   bool
	carryValid   bool
	usedDigits   []bool
	usedSymbols  map[string]struct{}
	labels       map[string]int
	addrLabels   []string
	outf         func(string, ...interface{})
	lastLogDump  int
}

func newGoGen(prob *planProblem) *goGen {
	return &goGen{
		planProblem: prob,
		usedSymbols: make(map[string]struct{}, 3*len(prob.letterSet)),
		usedDigits:  make([]bool, prob.base),
	}
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
	if i >= len(gg.addrLabels) {
		return ""
	}
	label := gg.addrLabels[i]
	if len(label) == 0 {
		return ""
	}
	return label
}

func (gg *goGen) logf(format string, args ...interface{}) {
	var label string
	if gg.addrLabels != nil {
		for _, arg := range args {
			if sol, ok := arg.(*solution); ok {
				label = gg.labelFor(sol.stepi)
				break
			}
		}
	}

	if label != "" {
		str := fmt.Sprintf(format, args...)
		if gg.outf == nil {
			log.Printf("%s  // %s", str, label)
		} else {
			gg.outf("%s  // %s", str, label)
		}
	} else if gg.outf == nil {
		log.Printf(format, args...)
	} else {
		gg.outf(format, args...)
	}
}

func (gg *goGen) init(desc string) {
}

func (gg *goGen) fix(c byte, v int) {
	gg.usedDigits[v] = true
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("fix(%s)", string(c))),
		setAStep(v),
		storeStep(c))
}

func (gg *goGen) saveCarry() {
	if gg.carryPrior != nil && !gg.carrySaved {
		if !gg.carryValid {
			panic("no valid carry to save")
		}
		gg.steps = append(gg.steps, setBAStep{})
		gg.carrySaved = true
	}
}

func (gg *goGen) computeSum(col *column) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for c:
	//   c = carry + a + b (mod base)
	a, b, c := col.cx[0], col.cx[1], col.cx[2]
	gg.ensurePriorCarry(col)
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("computeSum(%s)", charsLabel(a, b, c))))
	gg.saveCarry()
	gg.carryValid = false
	steps := make([]solutionStep, 0, 6)
	if a != 0 {
		steps = append(steps, addValueStep(a))
	}
	if b != 0 {
		steps = append(steps, addValueStep(b))
	}
	steps = append(steps,
		modStep(gg.base),
		storeStep(c))
	if c == gg.words[0][0] || c == gg.words[1][0] || c == gg.words[2][0] {
		steps = append(steps,
			relJNZStep(1),
			exitStep{errCheckFailed})
	}
	gg.steps = append(gg.steps, steps...)
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
	gg.ensurePriorCarry(col)
	gg.steps = append(gg.steps,
		labelStep(gg.gensym("computeSummand(%s)", charsLabel(a, b, c))))
	gg.saveCarry()
	gg.carryValid = false
	steps := make([]solutionStep, 0, 7)
	steps = append(steps, negateStep{})
	if c != 0 {
		steps = append(steps, addValueStep(c))
	}
	if b != 0 {
		steps = append(steps, subValueStep(b))
	}
	steps = append(steps,
		modStep(gg.base),
		storeStep(a))
	if a == gg.words[0][0] || a == gg.words[1][0] || a == gg.words[2][0] {
		steps = append(steps,
			relJNZStep(1),
			exitStep{errCheckFailed})
	}
	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) choose(c byte) {
	gg.saveCarry()
	gg.carryValid = false
	min := 0
	if gg.usedDigits[0] ||
		c == gg.words[0][0] ||
		c == gg.words[1][0] ||
		c == gg.words[2][0] {
		min = 1
	}

	var last = gg.base - 1
	for last > 0 && gg.usedDigits[last] {
		last--
	}
	for min <= last && gg.usedDigits[min] {
		min++
	}

	if min > last {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("no_choices_for(%s)", string(c))),
			exitStep{errNoChoices})
		return
	} else if min == last {
		gg.usedDigits[min] = true
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("only_choice_for(%s)", string(c))),
			setAStep(min),
			storeStep(c))
		return
	}

	if gg.useForkUntil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("choose(%s)", string(c))), // :choose($c)
			setAStep(min),                                 // ra = $min
			forkUntilStep(last),                           // forUntil $last
			storeStep(c),                                  // store $c
		)
	} else {
		var (
			loopSym     = gg.gensym("choose(%s):loop", string(c))
			nextLoopSym = gg.gensym("choose(%s):nextLoop", string(c))
			contSym     = gg.gensym("choose(%s):cont", string(c))
		)
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("choose(%s)", string(c))), // :choose($c)
			setAStep(min),                                 // ra = $min
			setCAStep{},                                   // rc = ra
			labelStep(loopSym),                            // :loop
			setACStep{},                                   // ra = rc
			isUsedStep{},                                  // used?
			labelJNZStep(nextLoopSym),                     // jnz :next_loop
			forkLabelStep(nextLoopSym),                    // fork :next_loop
			setACStep{},                                   // ra = rc
			labelJmpStep(contSym),                         // jmp :cont
			labelStep(nextLoopSym),                        // :nextLoop
			setACStep{},                                   // ra = rc
			addStep(1),                                    // add 1
			setCAStep{},                                   // rc = ra
			ltStep(last),                                  // lt $last
			labelJNZStep(loopSym),                         // jnz :loop
			setACStep{},                                   // ra = rc
			isUsedStep{},                                  // used?
			labelJZStep(contSym),                          // jz :cont
			exitStep{errAlreadyUsed},                      // exit errAlreadyUsed
			labelStep(contSym),                            // :cont
			setACStep{},                                   // ra = rc
			storeStep(c),                                  // store $c
		)
	}
}

func (gg *goGen) ensurePriorCarry(col *column) {
	pri := col.prior
	if pri == nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("ensureCarry(%d):noPrior", col.i)),
			setAStep(0))
	} else if pri == gg.carryPrior {
		if gg.carryValid {
			return
		} else if !gg.carrySaved {
			panic("no saved carry to restore")
		}
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("ensureCarry(%d):restore", col.i)),
			setABStep{})
		gg.carryValid = true
		return
	} else {
		c1, c2 := pri.cx[0], pri.cx[1]
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("ensureCarry(%d):compute(%s)", col.i, charsLabel(c1, c2))))
		gg.ensurePriorCarry(pri)
		steps := make([]solutionStep, 0, 3)
		if c1 != 0 {
			steps = append(steps, addValueStep(c1))
		}
		if c2 != 0 {
			steps = append(steps, addValueStep(c2))
		}
		steps = append(steps, divStep(gg.base))
		gg.steps = append(gg.steps, steps...)
	}
	gg.carryPrior = pri
	gg.carrySaved = false
	gg.carryValid = true
}

func (gg *goGen) checkColumn(col *column) {
	a, b, c := col.cx[0], col.cx[1], col.cx[2]
	gg.ensurePriorCarry(col)
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

func (gg *goGen) verify() {
	gg.steps = append(gg.steps, labelStep(gg.gensym("verify")))
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

	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) finish() {
	if gg.verified {
		gg.verify()
	}
	gg.steps = append(gg.steps, exitStep{nil})

	gg.labels = extractLabels(gg.steps, nil)
	gg.steps, gg.labels = eraseLabels(gg.steps, gg.labels)
	gg.steps, gg.labels = resolveLabels(gg.steps, gg.labels)

	gg.addrLabels = make([]string, len(gg.steps))
	for label, addr := range gg.labels {
		gg.addrLabels[addr] = label
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
