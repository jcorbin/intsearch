package main

import (
	"errors"
	"fmt"
	"log"
)

var (
	errAlreadyUsed    = errors.New("value already used")
	errCheckFailed    = errors.New("check failed")
	errNegativeValue  = errors.New("negative valued character")
	errDuplicateValue = errors.New("duplicate valued character")
	errVerifyFailed   = errors.New("verify failed")
)

type goGen struct {
	steps        []solutionStep
	debugLabels  bool
	verified     bool
	useForkUntil bool
	carrySaved   bool
	carryValid   bool
	usedSymbols  map[string]struct{}
	labels       map[string]int
	addrLabels   []string
	outf         func(string, ...interface{})
}

func (gg *goGen) obsAfter() afterGen {
	i := 0
	return afterGen(func(plan planner) {
		j := i
		for ; j < len(gg.steps); j++ {
			fmt.Printf("%v: %v\n", j, gg.steps[j])
		}
		if j > i {
			fmt.Println()
			i = j
		}
	})
}

func (gg *goGen) labelFor(sol *solution) string {
	if sol.stepi >= len(gg.addrLabels) {
		return ""
	}
	label := gg.addrLabels[sol.stepi]
	if len(label) == 0 {
		return ""
	}
	return label
}

func (gg *goGen) logf(format string, args ...interface{}) {
	var (
		sol   *solution
		label string
	)

	if gg.addrLabels != nil {
		for _, arg := range args {
			if s, ok := arg.(*solution); ok {
				sol = s
				break
			}
		}
	}

	if sol != nil {
		label = gg.labelFor(sol)
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

func (gg *goGen) init(plan planner, desc string) {
	prob := plan.problem()
	if len(gg.steps) > 0 {
		gg.steps = gg.steps[:0]
	}
	gg.usedSymbols = make(map[string]struct{}, 3*len(prob.letterSet))
	gg.labels = nil
	gg.addrLabels = nil
}

func (gg *goGen) setCarry(plan planner, v int) {
	if gg.debugLabels {
		gg.steps = append(gg.steps, labelStep(gg.gensym("setCarry")))
	}
	gg.steps = append(gg.steps, setAStep(v))
	gg.carrySaved = false
	gg.carryValid = true
}

func (gg *goGen) fix(plan planner, c byte, v int) {
	if gg.debugLabels {
		gg.steps = append(gg.steps, labelStep(gg.gensym(fmt.Sprintf("fix(%s)", string(c)))))
	}
	gg.steps = append(gg.steps,
		setAStep(v),
		storeStep(c))
}

func (gg *goGen) saveCarry(plan planner) {
	if !gg.carrySaved {
		if !gg.carryValid {
			panic("no valid carry to save")
		}
		gg.steps = append(gg.steps, setBAStep{})
		gg.carrySaved = true
	}
}

func (gg *goGen) restoreCarry(plan planner) {
	if !gg.carryValid {
		if !gg.carrySaved {
			panic("no saved carry to restore")
		}
		gg.steps = append(gg.steps, setABStep{})
		gg.carryValid = true
	}
}

func (gg *goGen) computeSum(plan planner, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for c:
	//   c = carry + a + b (mod base)
	if gg.debugLabels {
		label := gg.gensym(fmt.Sprintf("computeSum(%s, %s, %s)", string(a), string(b), string(c)))
		gg.steps = append(gg.steps, labelStep(label))
	}
	gg.restoreCarry(plan)
	gg.saveCarry(plan)
	gg.carryValid = false
	prob := plan.problem()
	steps := make([]solutionStep, 0, 6)
	if a != 0 {
		steps = append(steps, addValueStep(a))
	}
	if b != 0 {
		steps = append(steps, addValueStep(b))
	}
	steps = append(steps,
		modStep(prob.base),
		storeStep(c))
	if c == prob.words[0][0] || c == prob.words[1][0] || c == prob.words[2][0] {
		steps = append(steps,
			relJNZStep(1),
			exitStep{errCheckFailed})
	}
	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) computeSummand(plan planner, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for a:
	//   a = c - b - carry (mod base)
	if gg.debugLabels {
		label := gg.gensym(fmt.Sprintf("computeSummand(%s, %s, %s)", string(a), string(b), string(c)))
		gg.steps = append(gg.steps, labelStep(label))
	}
	gg.restoreCarry(plan)
	gg.saveCarry(plan)
	gg.carryValid = false
	prob := plan.problem()
	steps := make([]solutionStep, 0, 7)
	steps = append(steps, negateStep{})
	if c != 0 {
		steps = append(steps, addValueStep(c))
	}
	if b != 0 {
		steps = append(steps, subValueStep(b))
	}
	steps = append(steps,
		modStep(prob.base),
		storeStep(a))
	if a == prob.words[0][0] || a == prob.words[1][0] || a == prob.words[2][0] {
		steps = append(steps,
			relJNZStep(1),
			exitStep{errCheckFailed})
	}
	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) computeCarry(plan planner, c1, c2 byte) {
	if gg.debugLabels {
		label := gg.gensym(fmt.Sprintf("computeCarry(%s, %s)", string(c1), string(c2)))
		gg.steps = append(gg.steps, labelStep(label))
	}
	gg.restoreCarry(plan)
	prob := plan.problem()
	steps := make([]solutionStep, 0, 3)
	if c1 != 0 {
		steps = append(steps, addValueStep(c1))
	}
	if c2 != 0 {
		steps = append(steps, addValueStep(c2))
	}
	steps = append(steps, divStep(prob.base))
	gg.steps = append(gg.steps, steps...)
	gg.carryValid = true
	gg.carrySaved = false
}

func (gg *goGen) choose(plan planner, c byte) {
	if gg.debugLabels {
		label := gg.gensym(fmt.Sprintf("choose(%s)", string(c)))
		gg.steps = append(gg.steps, labelStep(label))
	}
	gg.saveCarry(plan)
	prob := plan.problem()
	steps := make([]solutionStep, 0, 22)
	gg.carryValid = false
	if c == prob.words[0][0] || c == prob.words[1][0] || c == prob.words[2][0] {
		steps = append(steps, setAStep(1))
	} else {
		steps = append(steps, setAStep(0))
	}
	var last = prob.base - 1
	if gg.useForkUntil {
		steps = append(steps, forkUntilStep(last))
	} else {
		var (
			loopSym     = gg.gensym(fmt.Sprintf("choose(%s):loop", string(c)))
			nextLoopSym = gg.gensym(fmt.Sprintf("choose(%s):nextLoop", string(c)))
			contSym     = gg.gensym(fmt.Sprintf("choose(%s):cont", string(c)))
		)
		steps = append(steps,
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
			ltStep(last),               // lt $last
			labelJNZStep(loopSym),      // jnz :loop
			setACStep{},                // ra = rc
			isUsedStep{},               // used?
			labelJZStep(contSym),       // jz :cont
			exitStep{errAlreadyUsed},   // exit errAlreadyUsed
			labelStep(contSym),         // :cont
			setACStep{},                // ra = rc
		)
	}
	steps = append(steps, storeStep(c))
	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) checkColumn(plan planner, cx [3]byte) {
	gg.restoreCarry(plan)
	steps := make([]solutionStep, 0, 9)

	n := 0
	if cx[0] != 0 {
		n++
		steps = append(steps, addValueStep(cx[0]))
	}
	if cx[1] != 0 {
		n++
		steps = append(steps, addValueStep(cx[1]))
	}
	if n > 0 {
		steps = append(steps,
			setCAStep{},
			modStep(prob.base))
	}
	steps = append(steps,
		subValueStep(cx[2]),
		relJZStep(1),
		exitStep{errCheckFailed})
	if n > 0 {
		steps = append(steps,
			setACStep{},
			divStep(prob.base))
	} else {
		steps = append(steps, setAStep(0))
	}
	gg.carrySaved = false
	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) verify(plan planner) {
	prob := plan.problem()

	N := len(prob.letterSet)
	C := prob.numColumns()
	steps := make([]solutionStep, 0, N*N/2*4+N*4+1+C*9+2)

	letters := make([]byte, 0, N)
	for c := range prob.letterSet {
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
		cx := prob.getColumn(i)
		if cx[0] != 0 {
			steps = append(steps, addValueStep(cx[0]))
		}
		if cx[1] != 0 {
			steps = append(steps, addValueStep(cx[1]))
		}
		steps = append(steps,
			setBAStep{},
			modStep(prob.base),
			subValueStep(cx[2]),
			relJZStep(1),
			exitStep{errVerifyFailed},
			setABStep{},
			divStep(prob.base))
	}
	steps = append(steps,
		relJZStep(1),
		exitStep{errVerifyFailed})

	gg.steps = append(gg.steps, steps...)
}

func (gg *goGen) finish(plan planner) {
	if gg.verified {
		gg.verify(plan)
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

func (gg *goGen) gensym(name string) string {
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
