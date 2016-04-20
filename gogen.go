package main

import (
	"errors"
	"fmt"
)

var (
	errAlreadyUsed  = errors.New("value already used")
	errCheckFailed  = errors.New("check failed")
	errVerifyFailed = errors.New("verify failed")
)

type goGen struct {
	steps        []solutionStep
	verified     bool
	useForkUntil bool
	carrySaved   bool
	carryValid   bool
	usedSymbols  map[string]struct{}
}

func (gg *goGen) obsAfter() *afterGen {
	i := 0
	return &afterGen{func(plan planner) {
		j := i
		for ; j < len(gg.steps); j++ {
			fmt.Printf("%v: %v\n", j, gg.steps[j])
		}
		if j > i {
			fmt.Println()
			i = j
		}
	}}
}

func (gg *goGen) init(plan planner, desc string) {
	prob := plan.problem()
	if len(gg.steps) > 0 {
		gg.steps = gg.steps[:0]
	}
	gg.usedSymbols = make(map[string]struct{}, 3*len(prob.letterSet))
}

func (gg *goGen) setCarry(plan planner, v int) {
	gg.steps = append(gg.steps, setAStep(v))
	gg.carrySaved = false
	gg.carryValid = true
}

func (gg *goGen) fix(plan planner, c byte, v int) {
	gg.steps = append(gg.steps, []solutionStep{
		setAStep(v),
		storeStep(c),
	}...)
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
	gg.restoreCarry(plan)
	gg.saveCarry(plan)
	gg.carryValid = false
	prob := plan.problem()
	if a != 0 {
		gg.steps = append(gg.steps, addValueStep(a))
	}
	if b != 0 {
		gg.steps = append(gg.steps, addValueStep(b))
	}
	gg.steps = append(gg.steps, []solutionStep{
		modStep(prob.base),
		storeStep(c),
	}...)
	if c == prob.words[0][0] || c == prob.words[1][0] || c == prob.words[2][0] {
		gg.steps = append(gg.steps, []solutionStep{
			relJNZStep(1),
			exitStep{errCheckFailed},
		}...)
	}
}

func (gg *goGen) computeSummand(plan planner, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for a:
	//   a = c - b - carry (mod base)
	gg.restoreCarry(plan)
	gg.saveCarry(plan)
	gg.carryValid = false
	prob := plan.problem()
	gg.steps = append(gg.steps, negateStep{})
	if c != 0 {
		gg.steps = append(gg.steps, addValueStep(c))
	}
	if b != 0 {
		gg.steps = append(gg.steps, subValueStep(b))
	}
	gg.steps = append(gg.steps, []solutionStep{
		modStep(prob.base),
		storeStep(a),
	}...)
	if a == prob.words[0][0] || a == prob.words[1][0] || a == prob.words[2][0] {
		gg.steps = append(gg.steps, []solutionStep{
			relJNZStep(1),
			exitStep{errCheckFailed},
		}...)
	}
}

func (gg *goGen) computeCarry(plan planner, c1, c2 byte) {
	gg.restoreCarry(plan)
	prob := plan.problem()
	if c1 != 0 {
		gg.steps = append(gg.steps, addValueStep(c1))
	}
	if c2 != 0 {
		gg.steps = append(gg.steps, addValueStep(c2))
	}
	gg.steps = append(gg.steps, divStep(prob.base))
	gg.carryValid = true
	gg.carrySaved = false
}

func (gg *goGen) gensym(name string) string {
	i := 1
	for {
		sym := fmt.Sprintf("%s(%d)", name, i)
		if _, used := gg.usedSymbols[sym]; !used {
			gg.usedSymbols[sym] = struct{}{}
			return sym
		}
		i++
	}
}

func (gg *goGen) choose(plan planner, c byte) {
	gg.saveCarry(plan)
	prob := plan.problem()
	gg.carryValid = false
	if c == prob.words[0][0] || c == prob.words[1][0] || c == prob.words[2][0] {
		gg.steps = append(gg.steps, setAStep(1))
	} else {
		gg.steps = append(gg.steps, setAStep(0))
	}
	var last = prob.base - 1
	if gg.useForkUntil {
		gg.steps = append(gg.steps, forkUntilStep(last))
	} else {
		var (
			loopSym     = gg.gensym(fmt.Sprintf("choose(%s):loop", string(c)))
			nextLoopSym = gg.gensym(fmt.Sprintf("choose(%s):nextLoop", string(c)))
			contSym     = gg.gensym(fmt.Sprintf("choose(%s):cont", string(c)))
		)
		gg.steps = append(gg.steps, []solutionStep{
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
		}...)
	}
	gg.steps = append(gg.steps, storeStep(c))
}

func (gg *goGen) checkFinal(plan planner, c byte, c1, c2 byte) {
	gg.restoreCarry(plan)
	gg.steps = append(gg.steps, []solutionStep{
		subValueStep(c),
		relJZStep(1),
		exitStep{errCheckFailed},
	}...)
}

func (gg *goGen) verify(plan planner) {
	prob := plan.problem()
	gg.steps = append(gg.steps, setAStep(0))
	for i := prob.numColumns() - 1; i >= 0; i-- {
		cx := prob.getColumn(i)
		if cx[0] != 0 {
			gg.steps = append(gg.steps, addValueStep(cx[0]))
		}
		if cx[1] != 0 {
			gg.steps = append(gg.steps, addValueStep(cx[1]))
		}
		gg.steps = append(gg.steps, []solutionStep{
			setBAStep{},
			modStep(prob.base),
			subValueStep(cx[2]),
			relJZStep(1),
			exitStep{errVerifyFailed},
			setABStep{},
			divStep(prob.base),
		}...)
	}
	gg.steps = append(gg.steps, []solutionStep{
		relJZStep(1),
		exitStep{errVerifyFailed},
	}...)
}

func (gg *goGen) finish(plan planner) {
	if gg.verified {
		gg.verify(plan)
	}
	gg.steps = append(gg.steps, exitStep{nil})
	labels := extractLabels(gg.steps, nil)
	gg.steps, labels = eraseLabels(gg.steps, labels)
	gg.steps, labels = resolveLabels(gg.steps, labels)
}
