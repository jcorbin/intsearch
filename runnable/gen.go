package runnable

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jcorbin/intsearch/word"
)

var (
	errAlreadyUsed = errors.New("value already used")
	errCheckFailed = errors.New("check failed")
	errNoChoices   = errors.New("no choices left")
)

// VerifyError is the error returned if final verification fails.
type VerifyError string

func (ve VerifyError) Error() string {
	return fmt.Sprintf("verify failed: %s", string(ve))
}

// StepGen implements a word.SolutionGen that builds a program listing of Steps
// that will progress the solution state.
type StepGen struct {
	*word.PlanProblem
	steps       []Step
	carryPrior  *word.Column
	carrySaved  bool
	carryValid  bool
	usedSymbols map[string]struct{}
	labels      map[string]int
	addrAnnos   map[int][]string
}

// NewStepGen creates a new step generator for a given problem about to be planned.
func NewStepGen(prob *word.PlanProblem) *StepGen {
	n := 0
	for _, w := range prob.Words {
		n += len(w)
	}
	gg := &StepGen{
		PlanProblem: prob,
		usedSymbols: make(map[string]struct{}, 3*len(prob.Letters)),
		steps:       make([]Step, 0, n*50),
	}
	if prob.Annotated {
		gg.addrAnnos = make(map[int][]string)
	}
	return gg
}

// Steps returns the slice of steps generated/compiled so far.
func (gg *StepGen) Steps() []Step {
	return gg.steps
}

func (gg *StepGen) copy() *StepGen {
	alt := &StepGen{
		PlanProblem: gg.PlanProblem,
		usedSymbols: gg.usedSymbols,
		steps:       make([]Step, 0, cap(gg.steps)),
	}
	// TODO: carry state copy... but whither column
	return alt
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

// SearchInit initializes the frontier with a solution with:
// - zero state
// - the planned problem
// - the compiled steps from the plan
// - the passed emit function for use by future fork and branch steps
//
// The returned sanity limit is based on the pessimistic assumption that every
// step will be ran a factorial number of times (a pathological brute force).
func (gg *StepGen) SearchInit(emit EmitFunc) int {
	emit(newSolution(&gg.PlanProblem.Problem, gg.steps, emit))
	// worst case, we have to run every step for every possible brute force solution
	numBrute := fallFact(gg.Base, len(gg.Letters))
	return numBrute * len(gg.steps)
}

// LoggedGen is a convenience that will create a new word.LogGen for the same
// problem, and wrap bundle it up with the StepGen into a word.MultiGen.
func (gg *StepGen) LoggedGen() word.SolutionGen {
	return word.MultiGen([]word.SolutionGen{
		word.NewLogGen(gg.PlanProblem),
		gg,
	})
}

// LabelAt returns any annotations for the given address, joined by a ", ".
func (gg *StepGen) LabelAt(i int) string {
	labels := gg.annosFor(i)
	if len(labels) == 0 {
		return ""
	}
	return strings.Join(labels, ", ")
}

// LabelFor returns any annotations for the given solution's current step,
// joined by a ", ".
func (gg *StepGen) LabelFor(sol *Solution) string {
	return gg.LabelAt(sol.stepi)
}

func (gg *StepGen) annosFor(addr int) []string {
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

// Decorate returns a list of any annotations knows for a any of the Solution
// arguments.
func (gg *StepGen) Decorate(args []interface{}) []string {
	if gg.addrAnnos == nil {
		return nil
	}
	var dec []string
	for _, arg := range args {
		if sol, ok := arg.(*Solution); ok {
			dec = append(dec, gg.annosFor(sol.stepi)...)
		}
	}
	return dec
}

// Logf does nothing.
func (gg *StepGen) Logf(format string, args ...interface{}) error {
	return nil
}

// Init does nothing.
func (gg *StepGen) Init(desc string) {
}

func (gg *StepGen) Fork(prob *word.PlanProblem, name, altLabel, contLabel string) word.SolutionGen {
	if altLabel != "" {
		altLabel = gg.gensym("%s:alt", altLabel)
	}
	if contLabel != "" {
		contLabel = gg.gensym("%s:cont", contLabel)
	}
	alt := gg.copy()
	alt.PlanProblem = prob
	gg.steps = append(gg.steps, forkAltStep{
		alt:       alt,
		name:      name,
		altLabel:  altLabel,
		contLabel: contLabel,
	})
	return alt
}

func (gg *StepGen) Fix(c byte, v int) {
	if gg.addrAnnos != nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("fix(%s)", string(c))))
	}
	gg.steps = append(gg.steps,
		setAStep(v),
		storeAStep(c))
}

func (gg *StepGen) stashCarry(col *word.Column) {
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

	if gg.addrAnnos != nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("stashCarry(%d)", col.I)))
	}
	gg.steps = append(gg.steps, setCAStep{})
	gg.carrySaved = true
	gg.carryPrior = col
}

func (gg *StepGen) saveCarry(col *word.Column) {
	if !gg.carryValid {
		gg.ensureCarry(col)
	}
	gg.stashCarry(col)
}

func (gg *StepGen) ComputeSum(col *word.Column) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for c:
	//   c = carry + a + b (mod base)
	a, b, c := col.Chars[0], col.Chars[1], col.Chars[2]
	gg.ensureCarry(col.Prior)
	gg.carryValid = false
	gg.carrySaved = false

	steps := make([]Step, 0, 12)
	if gg.addrAnnos != nil {
		steps = append(steps,
			labelStep(gg.gensym("computeSum(%s)", col.Label())))
	}
	if a != 0 {
		steps = append(steps, addAValueStep(a))
	}
	if b != 0 {
		steps = append(steps, addAValueStep(b))
	}
	steps = append(steps,
		setCAStep{},
		modAStep(gg.Base),
		setBAStep{},
		usedAStep{},
		relJZStep(1),
		exitStep{errCheckFailed},
		storeBStep(c),
		setACStep{},
		divAStep(gg.Base))
	gg.steps = append(gg.steps, steps...)

	gg.carryPrior = col
	gg.carryValid = true
	gg.carrySaved = false

	gg.checkAfterCompute(col, c)
}

func (gg *StepGen) ComputeFirstSummand(col *word.Column) {
	gg.computeSummand(col, col.Chars[0], col.Chars[1], col.Chars[2])
}

func (gg *StepGen) ComputeSecondSummand(col *word.Column) {
	gg.computeSummand(col, col.Chars[1], col.Chars[0], col.Chars[2])
}

func (gg *StepGen) computeSummand(col *word.Column, a, b, c byte) {
	// Given:
	//   carry + a + b = c (mod base)
	// Solve for a:
	//   a = c - b - carry (mod base)
	gg.ensureCarry(col.Prior)
	if !gg.carrySaved {
		gg.steps = append(gg.steps, setCAStep{})
	}

	gg.carryValid = false
	gg.carrySaved = false

	steps := make([]Step, 0, 10)
	if gg.addrAnnos != nil {
		steps = append(steps,
			labelStep(gg.gensym("computeSummand(%s)", col.Label())))
	}
	steps = append(steps, negAStep{})
	if c != 0 {
		steps = append(steps, addAValueStep(c))
	}
	if b != 0 {
		steps = append(steps, subAValueStep(b))
	}
	steps = append(steps,
		modAStep(gg.Base),
		setBAStep{},
		usedBStep{},
		relJZStep(1),
		exitStep{errCheckFailed},
		storeBStep(a))
	gg.steps = append(gg.steps, steps...)

	gg.carryPrior = col
	gg.carryValid = false
	gg.carrySaved = false

	if b != 0 {
		gg.steps = append(gg.steps,
			setACStep{},
			addARegBStep{},
			addAValueStep(b),
			divAStep(gg.Base),
		)
		gg.carryValid = true
	} else {
		gg.steps = append(gg.steps,
			setACStep{},
			addARegBStep{},
			divAStep(gg.Base),
		)
		gg.carryValid = true
	}

	gg.checkAfterCompute(col, a)
}

func (gg *StepGen) checkAfterCompute(col *word.Column, c byte) {
	if c == gg.Words[0][0] || c == gg.Words[1][0] || c == gg.Words[2][0] {
		gg.checkInitialLetter(col, c)
	}
	gg.checkFixedCarry(col)
}

func (gg *StepGen) checkInitialLetter(col *word.Column, c byte) {
	if gg.carryValid {
		gg.stashCarry(col)
		gg.carryValid = false
	}
	if gg.addrAnnos != nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("checkInitialLetter(%s)", string(c))))
	}
	gg.steps = append(gg.steps,
		loadAStep(c),
		relJNZStep(1),
		exitStep{errCheckFailed})
}

func (gg *StepGen) checkFixedCarry(col *word.Column) {
	switch col.Carry {
	case word.CarryZero:
		fallthrough
	case word.CarryOne:
		if !gg.restoreCarry(col) {
			return
		}
	default:
		return
	}

	if gg.addrAnnos != nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("checkFixedCarry(%s)", col.Label())))
	}

	switch col.Carry {
	case word.CarryZero:
		gg.steps = append(gg.steps,
			relJZStep(1),
			exitStep{errCheckFailed})
	case word.CarryOne:
		gg.steps = append(gg.steps,
			relJNZStep(1),
			exitStep{errCheckFailed})
	}
}

func (gg *StepGen) ChooseRange(c byte, min, max int) {
	gg.stashCarry(gg.carryPrior)
	gg.carryValid = false
	label := ""
	if gg.addrAnnos != nil {
		label = gg.gensym("choose(%s)", string(c))
	}
	gg.steps = append(gg.steps,
		rangeStep{label, min, max}, // range [$min, $max]
		storeBStep(c),              // store $c, rb
	)
}

func (gg *StepGen) restoreCarry(col *word.Column) bool {
	if col != gg.carryPrior {
		return false
	}
	if gg.carryValid {
		return true
	}
	if !gg.carrySaved {
		return false
	}
	if gg.addrAnnos != nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("restoreCarry(%d)", col.I)))
	}
	gg.steps = append(gg.steps, setACStep{})
	gg.carryValid = true
	return true
}

func (gg *StepGen) ensureCarry(col *word.Column) {
	if col == nil {
		if gg.addrAnnos != nil {
			gg.steps = append(gg.steps,
				labelStep(gg.gensym("ensureCarry:nil")))
		}
		gg.steps = append(gg.steps, setAStep(0))
		gg.carryPrior = nil
		gg.carrySaved = false
		gg.carryValid = true
		return
	}

	switch col.Carry {
	case word.CarryZero:
		fallthrough
	case word.CarryOne:
		if gg.addrAnnos != nil {
			gg.steps = append(gg.steps,
				labelStep(gg.gensym("ensureCarry(%d):fixed", col.I)))
		}
		gg.steps = append(gg.steps, setAStep(col.Carry))
		gg.carryPrior = col
		gg.carrySaved = false
		gg.carryValid = true
		return
	}

	if gg.restoreCarry(col) {
		return
	}

	c1 := col.Chars[0]
	if c1 != 0 && !gg.Known[c1] {
		log.Fatalf("cannot compute carry from unknown c1 for column %v", col)
	}

	c2 := col.Chars[1]
	if c2 != 0 && !gg.Known[c2] {
		log.Fatalf("cannot compute carry from unknown c2 for column %v", col)
	}

	gg.ensureCarry(col.Prior)
	if gg.addrAnnos != nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("computeCarry(%s)", col.Label())))
	}
	steps := make([]Step, 0, 3)
	if c1 != 0 {
		steps = append(steps, addAValueStep(c1))
	}
	if c2 != 0 {
		steps = append(steps, addAValueStep(c2))
	}
	steps = append(steps, divAStep(gg.Base))
	gg.steps = append(gg.steps, steps...)

	gg.carryPrior = col
	gg.carrySaved = false
	gg.carryValid = true
}

func (gg *StepGen) Check(err error) {
	if gg.addrAnnos == nil {
		gg.doVerify("", err)
	} else {
		gg.doVerify("check", err)
	}
}

func (gg *StepGen) CheckColumn(col *word.Column, err error) {
	if err == nil {
		err = errCheckFailed
	}

	a, b, c := col.Chars[0], col.Chars[1], col.Chars[2]
	gg.ensureCarry(col.Prior)
	if gg.addrAnnos != nil {
		gg.steps = append(gg.steps,
			labelStep(gg.gensym("checkColumn(%s)", col.Label())))
	}
	steps := make([]Step, 0, 9)

	n := 0
	if a != 0 {
		n++
		steps = append(steps, addAValueStep(a))
	}
	if b != 0 {
		n++
		steps = append(steps, addAValueStep(b))
	}
	if n > 0 {
		steps = append(steps,
			setCAStep{},
			modAStep(gg.Base))
	}
	steps = append(steps,
		subAValueStep(c),
		relJZStep(1),
		exitStep{err})
	if n > 0 {
		steps = append(steps,
			setACStep{},
			divAStep(gg.Base))
	} else {
		steps = append(steps, setAStep(0))
	}
	gg.carryPrior = col
	gg.carrySaved = false
	gg.carryValid = true
	gg.steps = append(gg.steps, steps...)
}

func (gg *StepGen) Verify() {
	if gg.addrAnnos == nil {
		gg.doVerify("", nil)
	} else {
		gg.doVerify("verify", nil)
	}
}

func (gg *StepGen) doVerify(name string, err error) {
	if name != "" {
		name = gg.gensym(name)
	}
	gg.steps = append(gg.steps, labelStep(name))
	gg.verifyInitialLetters(name, err)
	gg.verifyDuplicateLetters(name, err)
	gg.verifyLettersNonNegative(name, err)
	gg.verifyColumns(name, err)
}

func (gg *StepGen) verifyColumns(name string, err error) {
	// verify columns from bottom up
	for i := len(gg.Columns) - 1; i >= 0; i-- {
		if gg.Columns[i].Unknown > 0 {
			return
		}
		col := &gg.Columns[i]
		colErr := err
		if colErr == nil {
			colErr = VerifyError(col.Label())
		}
		gg.CheckColumn(col, colErr)
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
		finErr = VerifyError("final carry must be 0")
	}
	gg.steps = append(gg.steps,
		relJZStep(1),
		exitStep{finErr})
}

func (gg *StepGen) verifyInitialLetters(name string, err error) {
	if err == nil {
		err = VerifyError("initial letter cannot be zero")
	}
	if name != "" {
		gg.steps = append(gg.steps, labelStep(gg.gensym("%s:initialLetters", name)))
	}
	for _, word := range gg.Words {
		gg.steps = append(gg.steps,
			loadAStep(word[0]),
			relJNZStep(1),
			exitStep{err})
	}
}

func (gg *StepGen) verifyDuplicateLetters(name string, err error) {
	if err == nil {
		err = VerifyError("duplicate valued character")
	}
	if name != "" {
		gg.steps = append(gg.steps, labelStep(gg.gensym("%s:duplicateLetters", name)))
	}
	letters := gg.SortedLetters()
	for i, c := range letters {
		if !gg.Known[c] {
			continue
		}
		for j, d := range letters {
			if !gg.Known[d] {
				continue
			}
			if j > i {
				gg.steps = append(gg.steps,
					loadAStep(c),
					subAValueStep(d),
					relJNZStep(1),
					exitStep{err})
			}
		}
	}
}

func (gg *StepGen) verifyLettersNonNegative(name string, err error) {
	if err == nil {
		err = VerifyError("negative valued character")
	}
	if name != "" {
		gg.steps = append(gg.steps, labelStep(gg.gensym("%s:allLettersNonNegative", name)))
	}
	for _, c := range gg.SortedLetters() {
		if !gg.Known[c] {
			continue
		}
		gg.steps = append(gg.steps,
			loadAStep(c),
			ltAStep(0),
			relJZStep(1),
			exitStep{err})
	}
}

func (gg *StepGen) Finish() {
	lastStep := gg.steps[len(gg.steps)-1]
	if _, isFinish := lastStep.(finishStep); isFinish {
		panic("double StepGen.finish")
	}
	gg.steps = append(gg.steps, finishStep(gg.gensym("finish")))
}

func (gg *StepGen) Finalize() {
	gg.compile()
}

func (gg *StepGen) takeAnnotation(addr int, annos ...string) {
	gg.addrAnnos[addr] = append(gg.addrAnnos[addr], annos...)
}

func (gg *StepGen) compile() {
	var parts [][]Step
	var addr int
	var annotate annoFunc
	if gg.addrAnnos != nil {
		annotate = gg.takeAnnotation
	}
	addr, parts, gg.labels = expandSteps(addr, gg.steps, nil, gg.labels, annotate)
	steps := make([]Step, 0, addr)
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

func (gg *StepGen) gensym(format string, args ...interface{}) string {
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

func printLastKSteps(k int, steps []Step) {
	i := len(steps) - k - 1
	if i < 0 {
		i = 0
	}
	for j, step := range steps[i:] {
		fmt.Printf("[%v]: %v\n", i+j, step)
	}
}
