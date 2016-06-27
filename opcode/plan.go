package opcode

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jcorbin/intsearch/internal"
	"github.com/jcorbin/intsearch/word"
)

type exitCode byte

const (
	codeNormal = exitCode(iota)
	codeAlreadyUsed
	codeCheckFailed
	codeNoChoices
	codeVerifyInitialLetters
	codeVerifyDuplicateLetters
	codeVerifyNegativeValue
	codeVerifyColumn
	codeVerifyFinalCarry
	codeCustom     = 0x80
	customCodeMask = 0x7f
)

var (
	errAlreadyUsed            = errors.New("value already used")
	errCheckFailed            = errors.New("check failed")
	errNoChoices              = errors.New("no choices left")
	errVerifyInitialLetters   = errors.New("initial letter cannot be zero")
	errVerifyDuplicateLetters = errors.New("duplicate valued character")
	errVerifyNegativeValue    = errors.New("negative valued character")
	errVerifyColumn           = errors.New("column addition failed")
	errVerifyFinalCarry       = errors.New("final carry must be 0")
)

func setExitCodeOp(code exitCode) Op {
	return MoveLOp(Location(0x0000), Immediate(uint16(code)))
}

var offExit = int16(setExitCodeOp(0).EncodedSize() + HaltOp.EncodedSize())

// Plan implements word.Plan around an opcode machine.
type Plan struct {
	*word.PlanProblem

	errors        [128]error
	errorCodes    map[error]exitCode
	nextErrorCode exitCode

	// memory offsets for values
	valOff [256]uint16

	// memory offsets for used flags; TODO try using a bit vector in a
	// dedicated register instead
	useOff [50]uint16

	mach  Machine
	annos map[int][]string
}

// Init initializes the plan.
func (plan *Plan) Init() {
	if len(plan.Letters) > 50 {
		panic("can't support more than 50 letters")
	}
	if plan.Base > 50 {
		panic("can't support more than base 50")
	}
	off := uint16(1)
	for _, c := range plan.SortedLetters() {
		plan.valOff[c] = off
		off += 2
	}
	for i := 0; i < plan.Base; i++ {
		plan.useOff[i] = off
		off += 2
	}
	plan.errorCodes = make(map[error]exitCode)
	if plan.Annotated {
		plan.annos = make(map[int][]string)
	}
}

// DefineError defines an error, if it's not already defined, and returns its
// code.
func (plan *Plan) DefineError(err error) exitCode {
	if code, defined := plan.errorCodes[err]; defined {
		return code
	}
	if plan.nextErrorCode == 0 {
		plan.nextErrorCode++
	}
	code := plan.nextErrorCode
	plan.nextErrorCode++
	plan.errorCodes[err] = code
	plan.errors[code] = err
	return code | codeCustom
}

// CodeFor returns the code for an error, and whether or not it was defined.
func (plan *Plan) CodeFor(err error) (code exitCode, defined bool) {
	switch err {
	case nil:
		return codeNormal, true
	case errAlreadyUsed:
		return codeAlreadyUsed, true
	case errCheckFailed:
		return codeCheckFailed, true
	case errNoChoices:
		return codeNoChoices, true
	case errVerifyInitialLetters:
		return codeVerifyInitialLetters, true
	case errVerifyDuplicateLetters:
		return codeVerifyDuplicateLetters, true
	case errVerifyNegativeValue:
		return codeVerifyNegativeValue, true
	case errVerifyColumn:
		return codeVerifyColumn, true
	case errVerifyFinalCarry:
		return codeVerifyFinalCarry, true
	}
	code, defined = plan.errorCodes[err]
	return
}

// ErrorFor returns the error for a code, or nil if no such code is defined.
func (plan *Plan) ErrorFor(code exitCode) error {
	switch code {
	case codeNormal:
		return nil
	case codeAlreadyUsed:
		return errAlreadyUsed
	case codeCheckFailed:
		return errCheckFailed
	case codeNoChoices:
		return errNoChoices
	case codeVerifyInitialLetters:
		return errVerifyInitialLetters
	case codeVerifyDuplicateLetters:
		return errVerifyDuplicateLetters
	case codeVerifyNegativeValue:
		return errVerifyNegativeValue
	case codeVerifyColumn:
		return errVerifyColumn
	case codeVerifyFinalCarry:
		return errVerifyFinalCarry
	}
	if code&codeCustom == 0 {
		return fmt.Errorf("unknown exit code %02x", byte(code))
	}
	code = code & customCodeMask
	if err := plan.errors[code]; err != nil {
		return err
	}
	return fmt.Errorf("unknown custom exit code %02x", byte(code))
}

// Problem returns the plan(ed/ing) word problem.
func (plan *Plan) Problem() *word.Problem {
	return &plan.PlanProblem.Problem
}

// Run runs the loaded program, and emits solutions derived from completed
// machines.
func (plan *Plan) Run(res word.Resultor) {
	if wat, ok := res.(word.Watcher); ok {
		plan.mach.RunAll(&planTracer{
			Plan: plan,
			wat:  wat,
			sols: make(map[Machine]*solution),
		})
	} else {
		plan.mach.RunAll(&planResultor{
			Plan: plan,
			res:  res,
		})
	}
}

// Dump prints the generated program and any annotations.
func (plan *Plan) Dump(logf func(format string, args ...interface{})) {
	p := plan.mach.Program()
	for i := 0; i < len(p); {
		op, j := DecodeOp(plan.mach.ByteOrder(), p, i)
		if annos := plan.annosFor(i); len(annos) > 0 {
			logf("%04x: %v // %s", i, op, strings.Join(annos, " "))
		} else {
			logf("%04x: %v", i, op)
		}
		i = j
	}
}

func (plan *Plan) annosFor(off int) []string {
	if plan.annos != nil {
		return plan.annos[off]
	}
	return nil
}

// Decorate returns any annotations avaliable for the given args.
func (plan *Plan) Decorate(args ...interface{}) []string {
	if plan.annos == nil {
		return nil
	}
	var annos []string
	for _, arg := range args {
		if annArg, ok := arg.(annotator); ok {
			annos = append(annos, annArg.Annotate(plan.annos)...)
		}
	}
	return annos
}

type planResultor struct {
	*Plan
	res word.Resultor
}

func (pr *planResultor) Result(mach Machine) bool {
	return pr.res.Result(&solution{
		Plan: pr.Plan,
		mach: mach,
	})
}

type planTracer struct {
	*Plan
	wat  word.Watcher
	sols map[Machine]*solution
}

func (pt *planTracer) solution(mach Machine) (sol *solution) {
	sol = pt.sols[mach]
	if sol == nil {
		sol = &solution{
			Plan: pt.Plan,
			mach: mach,
		}
		pt.sols[mach] = sol
	}
	return
}

func (pt *planTracer) Result(mach Machine) bool {
	sol := pt.solution(mach)
	sol.state = solutionResult
	delete(pt.sols, mach)
	return pt.wat.Result(sol)
}

func (pt *planTracer) Before(mach Machine) {
	sol := pt.solution(mach)
	sol.state = solutionBefore
	pt.wat.Before(sol)
}

func (pt *planTracer) After(mach Machine) {
	sol := pt.solution(mach)
	sol.state = solutionAfter
	pt.wat.After(sol)
}

func (pt *planTracer) Emit(action string, parent, child Machine) {
	var p word.Solution
	if parent != nil {
		sol := pt.solution(parent)
		sol.state = solutionForkCont
		sol.forkKind = action
		p = sol
	}
	sol := pt.solution(child)
	sol.state = solutionForkDefer
	sol.forkKind = action
	pt.wat.Fork(p, sol)
}

type annotator interface {
	Annotate(annos map[int][]string) []string
}

type solutionState int

const (
	solutionResult = solutionState(iota)
	solutionBefore
	solutionAfter
	solutionForkDefer
	solutionForkCont
)

type solution struct {
	*Plan
	mach     Machine
	state    solutionState
	forkKind string
}

type solutionLetVal struct {
	let   byte
	val   uint16
	valid bool
	known bool
}

func (lv solutionLetVal) String() string {
	if !lv.valid {
		return fmt.Sprintf("INVALID(%s=%04x)", string(lv.let), lv.val)
	}
	if !lv.known {
		return fmt.Sprintf("UNKNOWN(%s=%04x)", string(lv.let), lv.val)
	}
	return fmt.Sprintf("%s=%d", string(lv.let), lv.val)
}

func (sol *solution) letVal(c byte) (lv solutionLetVal) {
	lv.let = c
	var buf [2]byte
	i := sol.mach.CopyMemory(int(sol.valOff[c]), buf[:])
	lv.val = sol.mach.ByteOrder().Uint16(buf[:i])
	lv.known = lv.val != uint16(-int16(c))
	lv.valid = !lv.known || lv.val < uint16(sol.Base)
	return
}

func (sol *solution) ValueOf(c byte) (int, bool) {
	lv := sol.letVal(c)
	return int(lv.val), lv.known
}

func (sol *solution) Check() error {
	err := sol.mach.Check()
	if nz, ok := err.(ErrNonZeroHalt); ok {
		err = sol.ErrorFor(exitCode(nz))
	}
	return err
}

func (sol *solution) nextPI() int {
	return sol.mach.PI()
}

func (sol *solution) lastPI() int {
	return sol.mach.PI() - sol.mach.LastOp().EncodedSize()
}

func (sol *solution) Annotate(annos map[int][]string) []string {
	switch sol.state {
	case solutionBefore:
		return annos[sol.nextPI()]
	case solutionAfter:
		return annos[sol.lastPI()]
	case solutionForkDefer:
		fallthrough
	case solutionForkCont:
		return annos[sol.nextPI()]
	case solutionResult:
		return annos[sol.lastPI()]
	}
	panic("invalid solution.state")
}

func (sol *solution) String() string {
	switch sol.state {
	case solutionBefore:
		return fmt.Sprintf("%v // before: %v", sol.mach.State(), sol.mach.NextOp())

	case solutionAfter:
		return fmt.Sprintf("%v // after: %v", sol.mach.State(), sol.mach.LastOp())

	case solutionForkDefer:
		return fmt.Sprintf("%v // deferred: %v", sol.mach.State(), sol.mach.NextOp())

	case solutionForkCont:
		return fmt.Sprintf("%v // continue: %v", sol.mach.State(), sol.mach.NextOp())

	case solutionResult:
		return fmt.Sprintf("%v", sol.mach.State())

	default:
		panic("invalid solution.state")
	}
}

func (sol *solution) Dump(logf func(string, ...interface{})) {
	switch sol.state {
	case solutionBefore:
		sol.dumpBefore(logf)

	case solutionAfter:
		sol.dumpAfter(logf)

	case solutionForkDefer:
		sol.dumpFork("defer", logf)

	case solutionForkCont:
		sol.dumpFork("cont", logf)

	case solutionResult:
		sol.dumpResult(logf)

	default:
		panic("invalid solution.state")
	}
}

func (sol *solution) dumpBefore(logf func(string, ...interface{})) {
	logf("state: %v", sol)
}

func (sol *solution) dumpAfter(logf func(string, ...interface{})) {
	op := sol.mach.LastOp()
	logf("state: %v", sol)
	switch op.Code {
	case MOVE:
		fallthrough
	case MOVEL:
		fallthrough
	case MOVEH:
		if op.Arg1.Code.Indirect() {
			logf("mapping: %s", sol.letterValueString())
			dumpMemory(sol.mach, internal.ElidedF(logf, "memory:"))
		}
	}
}

func (sol *solution) dumpFork(which string, logf func(string, ...interface{})) {
	logf("%s %s: %v", which, sol.forkKind, sol)
}

func (sol *solution) dumpResult(logf func(string, ...interface{})) {
	err := sol.mach.Check()
	if nz, ok := err.(ErrNonZeroHalt); ok {
		err = sol.ErrorFor(exitCode(nz))
	}
	logf("final state: %v err=%v", sol, err)
	logf("final mapping: %s", sol.letterValueString())
	dumpMemory(sol.mach, internal.ElidedF(logf, "memory:"))
}

func (sol *solution) letterValueString() string {
	parts := make([]interface{}, 0, len(sol.Letters))
	for _, c := range sol.SortedLetters() {
		if lv := sol.letVal(c); lv.known || !lv.valid {
			parts = append(parts, lv)
		}
	}
	return fmt.Sprint(parts)
}
