package opcode

import (
	"fmt"

	"github.com/jcorbin/intsearch/word"
)

// CodeGen implements a word.SolutionGen that assembles a machine program that
// will solve a given word.PlanProblem
type CodeGen struct {
	Assembler
	regs RegisterAllocator
	Plan

	// the register names used for col.I's carry
	carryRegs []string

	opLimRef Ref
	annos    map[int][]string
	finished bool
}

// CodeGenFork is an alternate copy of a CodeGen.
type CodeGenFork struct {
	CodeGen
	parent    *CodeGen
	priorLen  int
	contLabel string
	contRef   Ref
}

// NewCodeGen creates a new opcode generator for a given problem about to be planned.
func NewCodeGen(prob *word.PlanProblem) word.SolutionGen {
	n := 0
	for _, w := range prob.Words {
		n += len(w)
	}
	sizeEst := n * 50 * 7 // budget 50 max-width operations for each (non-unique!) letter
	return &CodeGen{
		Assembler: Assembler{
			bo:  TinyMachineByteOrder(),
			buf: make([]byte, 0, sizeEst),
		},
		Plan: Plan{
			PlanProblem: prob,
		},
	}
}

// Label calls Assembler.Label, and adds an annotation if in annotated mode.
func (cg *CodeGen) Label(name string) {
	cg.Annotate(fmt.Sprintf(":%s", name))
}

// Annotate adds an annotation for the given address, if annotation are
// enabled.
func (cg *CodeGen) Annotate(anno string) {
	if cg.annos != nil {
		offset := len(cg.buf)
		cg.annos[offset] = append(cg.annos[offset], anno)
	}
}

// Annotatef adds a formatted annotation for the given address, if annotation
// are enabled.
func (cg *CodeGen) Annotatef(format string, args ...interface{}) {
	if cg.annos != nil {
		offset := len(cg.buf)
		cg.annos[offset] = append(cg.annos[offset],
			fmt.Sprintf(format, args...))
	}
}

// Problem returns the plan problem.
func (cg *CodeGen) Problem() *word.PlanProblem {
	return cg.PlanProblem
}

// Logf does nothing.
func (cg *CodeGen) Logf(format string, args ...interface{}) error {
	return nil
}

// Init computes memory offsets for the given program.
func (cg *CodeGen) Init(desc string) {
	cg.regs.Init(tinyMachineNumRegisters) // TODO: decouple from tiny machine
	cg.Plan.Init()
	cg.carryRegs = make([]string, len(cg.Columns))
	for i := range cg.Columns {
		cg.carryRegs[i] = fmt.Sprintf("C%d", i)
	}

	cg.opLimRef = cg.WriteOpRef(OpLim(0))
	for _, c := range cg.SortedLetters() {
		cg.WriteOp(MoveOp(Location(cg.valOff[c]), Immediate(uint16(-int16(c)))))
	}
}

// Fork creates a copy of the CodeGen such that:
// - a copy of execution context will execute the copy's code
// - while control flow carries on executing the original's code
//
// The copy's code immediately follows in program memory, followed by the rest
// of the original code.
//
// The caller:
// - MUST call Finish on the copy before calling any
//   other methods on the original.
// - MUST NOT call Finalize on the copy
// - MAY call Fork on the copy (and its copies to an arbitrary depth)
//
// Example:
//     var orig StepGen
//     copy := orig.Fork(prob, "name", "alt", "cont")
//     // ... call copy methods
//     copy.Finish()
//     // ... call orig methods
func (cg *CodeGen) Fork(prob *word.PlanProblem, name, altLabel, contLabel string) word.SolutionGen {
	if altLabel != "" {
		cg.Label(altLabel)
	}
	alt := &CodeGenFork{
		CodeGen:   *cg,
		parent:    cg,
		priorLen:  len(cg.buf),
		contLabel: contLabel,
	}
	alt.PlanProblem = prob
	alt.regs.Copy()
	alt.contRef = alt.WriteOpRef(BranchOp(0))
	return alt
}

// Fix adds code to fix the value of c to v.
func (cg *CodeGen) Fix(c byte, v int) {
	if cg.Annotated {
		cg.Label(fmt.Sprintf("fix(%s)", string(c)))
	}
	cg.WriteOp(MoveOp(Location(cg.valOff[c]), Immediate(uint16(v))))
	cg.WriteOp(MoveOp(Location(cg.useOff[v]), Immediate(1)))
}

// ComputeSum adds code to compute the sum character from its summands.
//
// That is, given a column:
//     carry + a + b = c (mod base)
// Compute c:
//     c = carry + a + b (mod base)
func (cg *CodeGen) ComputeSum(col *word.Column) {
	a, b, c := col.Chars[0], col.Chars[1], col.Chars[2]

	if cg.Annotated {
		cg.Label(fmt.Sprintf("computeSum(%s)", col.Label()))
	}

	// compute letter value
	rc := cg.setupCarryCompute(col)
	if a != 0 {
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rc, Location(cg.valOff[a])))
	}
	if b != 0 {
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rc, Location(cg.valOff[b])))
	}

	rl := Register(cg.regs.Take(string(c)))
	cg.Annotatef("%s=%s", string(c), cg.carryRegs[col.I])
	cg.WriteOp(MoveOp(rl, rc))
	cg.WriteOp(ModOp(rl, Immediate(uint16(cg.Base))))

	// store computed letter value
	useLoc := cg.checkUsed(rl)
	cg.WriteOp(MoveOp(Location(cg.valOff[c]), rl))
	cg.WriteOp(MoveOp(useLoc, Immediate(1)))

	// compute carry
	cg.WriteOp(DivOp(rc, Immediate(uint16(cg.Base))))

	cg.checkAfterCompute(col, c, rl, rc)
}

// ComputeFirstSummand adds code to compute the first summand from its sum and
// the other summand.
//
// That is, given a column:
//     carry + a + b = c (mod base)
// Compute a:
//     a = c - b - carry (mod base)
func (cg *CodeGen) ComputeFirstSummand(col *word.Column) {
	cg.computeSummand(col, col.Chars[0], col.Chars[1], col.Chars[2])
}

// ComputeSecondSummand adds code to compute the second summand from its sum
// and the other summand.
//
// That is, given a column:
//     carry + a + b = c (mod base)
// Compute b:
//     b = c - a - carry (mod base)
func (cg *CodeGen) ComputeSecondSummand(col *word.Column) {
	cg.computeSummand(col, col.Chars[1], col.Chars[0], col.Chars[2])
}

// computeSummand adds code to solve for `a` in `carry + a + b = c (mod base)`
//
// To do so the computation is:
//     a = c - b - carry (mod base)
// In particular:
//     a = (-carry + c - b) % base
func (cg *CodeGen) computeSummand(col *word.Column, a, b, c byte) {
	if cg.Annotated {
		cg.Label(fmt.Sprintf("computeSummand(%s, %s)", string(a), col.Label()))
	}

	// compute letter value
	rc := cg.setupCarryCompute(col)
	rl := Register(cg.regs.Take(string(a)))

	cg.Annotatef("%s=%s", string(c), cg.carryRegs[col.I])
	cg.WriteOp(MoveOp(rl, rc))
	cg.WriteOp(NegOp(rl))
	if c != 0 {
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rl, Location(cg.valOff[c])))
	}
	if b != 0 {
		// TODO: re-use any letter register
		cg.WriteOp(SubOp(rl, Location(cg.valOff[b])))
	}

	cg.WriteOp(AddOp(rl, Immediate(uint16(cg.Base))))
	cg.WriteOp(ModOp(rl, Immediate(uint16(cg.Base))))

	// store computed letter value
	useLoc := cg.checkUsed(rl)
	cg.WriteOp(MoveOp(Location(cg.valOff[a]), rl))
	cg.WriteOp(MoveOp(useLoc, Immediate(1)))

	// compute `carry = priorCarry + a + b / base`
	if b != 0 {
		// TODO: consider combining with above b != 0 branch
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rc, Location(cg.valOff[b])))
	}
	cg.WriteOp(AddOp(rc, rl))
	cg.WriteOp(DivOp(rc, Immediate(uint16(cg.Base))))

	cg.checkAfterCompute(col, a, rl, rc)
}

// ChooseRange adds code to try all remaining unchosen digits in a range for
// the character c.
func (cg *CodeGen) ChooseRange(c byte, min, max int) {
	var label string
	if cg.Annotated {
		label = fmt.Sprintf("choose(%s, %d, %d)", string(c), min, max)
		cg.Label(label)
	}

	// try each value from min to max
	rl := Register(cg.regs.Take(string(c)))
	ri := Register(cg.regs.Take("index"))
	useLoc := Indexed(cg.useOff[0], ri.Code.Register())

	cg.WriteOp(MoveOp(rl, Immediate(uint16(min)))) // l = min

	bodyOffset := len(cg.buf)
	if cg.Annotated {
		cg.Label(fmt.Sprintf("%s:body", label))
	}

	cg.WriteOp(MoveOp(ri, rl))                   // ri = rl ...
	cg.WriteOp(MulOp(ri, Immediate(2)))          // ... / 2
	cg.WriteOp(EQOp(useLoc, Immediate(0)))       // ? l not used?
	bref := cg.WriteOpRef(BranchTOp(0))          // T *> cont
	cg.WriteOp(AddOp(rl, Immediate(1)))          // l++
	cg.WriteOp(LTOp(rl, Immediate(uint16(max)))) // ? l < max
	jref := cg.WriteOpRef(JumpTOp(0))            // T -> body
	cg.WriteOp(MoveOp(ri, rl))                   // ri = rl ...
	cg.WriteOp(MulOp(ri, Immediate(2)))          // ... / 2
	cg.WriteOp(EQOp(useLoc, Immediate(0)))       // ? l not used?
	cg.WriteOp(JumpTOp(offExit))                 // T -> skip exit
	cg.WriteOp(setExitCodeOp(codeAlreadyUsed))   // exit already used
	cg.WriteOp(HaltOp)                           // halt

	contOffset := len(cg.buf)
	if cg.Annotated {
		cg.Label(fmt.Sprintf("%s:cont", label))
	}

	cg.WriteOp(MoveOp(Location(cg.valOff[c]), rl)) // store value
	cg.WriteOp(MoveOp(useLoc, Immediate(1)))       // mark value used

	bref.WriteOffset(contOffset) // branch -> :cont
	jref.WriteOffset(bodyOffset) // jump -> :body (loop)
}

// CheckColumn adds code to check that a given column (carry + a + b == c);
// this is used for columns whose letters were all determined by other columns.
func (cg *CodeGen) CheckColumn(col *word.Column, err error) {
	if err == nil {
		err = errCheckFailed
	}
	errCode := cg.DefineError(err)
	if cg.Annotated {
		cg.checkColumn(col, errCode, "checkColumn")
	} else {
		cg.checkColumn(col, errCode, "")
	}
}

// checkColumn is the internal implementation of CheckColumn, also reused by
// verifyColumns.
func (cg *CodeGen) checkColumn(col *word.Column, errCode exitCode, name string) {
	a, b, c := col.Chars[0], col.Chars[1], col.Chars[2]
	if name != "" {
		cg.Label(fmt.Sprintf("%s(%s)", name, col.Label()))
	}

	// compute carry + a + b
	rc := cg.setupCarryCompute(col)

	n := 0
	if a != 0 {
		n++
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rc, Location(cg.valOff[a])))
	}
	if b != 0 {
		n++
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rc, Location(cg.valOff[b])))
	}

	rl := Register(cg.regs.Take(string(c)))
	cg.Annotatef("%s=%s", string(c), cg.carryRegs[col.I])
	cg.WriteOp(MoveOp(rl, rc))
	if n > 0 {
		cg.WriteOp(ModOp(rl, Immediate(uint16(cg.Base))))
	}

	// check against prior value
	cg.WriteOp(EQOp(rl, Location(cg.valOff[c])))
	cg.WriteOp(JumpTOp(offExit))
	cg.WriteOp(setExitCodeOp(errCode))
	cg.WriteOp(HaltOp)

	// compute `carry = priorCarry + a + b / base`
	if n > 0 {
		cg.WriteOp(DivOp(rc, Immediate(uint16(cg.Base))))
	} else {
		cg.Annotatef("%s=0", cg.carryRegs[col.I])
		cg.WriteOp(MoveOp(rc, Immediate(0)))
	}
}

// Check adds code that check if the selected mapping of letters to digits
// works.  Check's work failing is a normal terminal solution (e.g. used by
// brute force search).
func (cg *CodeGen) Check(err error) {
	if cg.Annotated {
		cg.doVerify("check", err)
	} else {
		cg.doVerify("", err)
	}
}

// Finish adds the final successful exit.
func (cg *CodeGen) Finish() {
	if cg.finished {
		panic("double CodeGen.finish")
	}
	cg.finished = true
	cg.WriteOp(setExitCodeOp(codeNormal))
	cg.WriteOp(HaltOp)
}

// Verify adds code that verifies that the selected mapping of letters to
// digits works.  Verify's work failing is an error terminal solution (e.g.
// used to test correctness of the planned program).
func (cg *CodeGen) Verify() {
	if cg.Annotated {
		cg.doVerify("verify", nil)
	} else {
		cg.doVerify("", nil)
	}
}

// doVerify add codes for Check and Verify.
func (cg *CodeGen) doVerify(name string, err error) {
	if name != "" {
		cg.Label(name)
	}
	if err == nil {
		cg.verifyInitialLetters(name, codeVerifyInitialLetters)
		cg.verifyDuplicateLetters(name, codeVerifyDuplicateLetters)
		cg.verifyLettersNonNegative(name, codeVerifyNegativeValue)
		cg.verifyColumns(name, codeVerifyColumn)
		cg.verifyFinalCarry(name, codeVerifyFinalCarry)
	} else {
		errCode := cg.DefineError(err)
		cg.verifyInitialLetters(name, errCode)
		cg.verifyDuplicateLetters(name, errCode)
		cg.verifyLettersNonNegative(name, errCode)
		cg.verifyColumns(name, errCode)
		cg.verifyFinalCarry(name, errCode)
	}
}

// verifyInitialLetters adds code to check that all initial letters are
// non-zero.
func (cg *CodeGen) verifyInitialLetters(name string, errCode exitCode) {
	if name != "" {
		cg.Label(fmt.Sprintf("%s:initialLetters", name))
	}
	for _, word := range cg.Words {
		cg.WriteOp(EQOp(Location(cg.valOff[word[0]]), Immediate(0)))
		cg.WriteOp(JumpFOp(offExit))
		cg.WriteOp(setExitCodeOp(errCode))
		cg.WriteOp(HaltOp)
	}
}

// verifyDuplicateLetters adds code to check that no two letters are set to the
// same value.
func (cg *CodeGen) verifyDuplicateLetters(name string, errCode exitCode) {
	if name != "" {
		cg.Label(fmt.Sprintf("%s:duplicateLetters", name))
	}
	letters := cg.SortedLetters()
	for i, c := range letters {
		if !cg.Known[c] {
			continue
		}
		for j, d := range letters {
			if !cg.Known[d] {
				continue
			}
			if j > i {
				cg.WriteOp(EQOp(Location(cg.valOff[c]), Location(cg.valOff[d])))
				cg.WriteOp(JumpFOp(offExit))
				cg.WriteOp(setExitCodeOp(errCode))
				cg.WriteOp(HaltOp)
			}
		}
	}
}

// verifyLettersNonNegative adds code to check that no known letter has been
// set to a negative value.
func (cg *CodeGen) verifyLettersNonNegative(name string, errCode exitCode) {
	if name != "" {
		cg.Label(fmt.Sprintf("%s:allLettersNonNegative", name))
	}
	for _, c := range cg.SortedLetters() {
		if !cg.Known[c] {
			continue
		}
		cg.WriteOp(LTOp(Location(cg.valOff[c]), Immediate(0)))
		cg.WriteOp(JumpFOp(offExit))
		cg.WriteOp(setExitCodeOp(errCode))
		cg.WriteOp(HaltOp)
	}
}

// verifyColumns adds code to check each column's addition.
func (cg *CodeGen) verifyColumns(name string, errCode exitCode) {
	for i := len(cg.Columns) - 1; i >= 0; i-- {
		if cg.Columns[i].Unknown > 0 {
			return
		}
		col := &cg.Columns[i]
		cg.checkColumn(col, errCode, name)
	}
}

// verifyFinalCarry adds code to check the final carry.
func (cg *CodeGen) verifyFinalCarry(name string, errCode exitCode) {
	if name != "" {
		cg.Label(fmt.Sprintf("%s:finalCarry", name))
	}
	rc := cg.carryArg(&cg.Columns[0])
	cg.WriteOp(EQOp(rc, Immediate(0)))
	cg.WriteOp(JumpTOp(offExit))
	cg.WriteOp(setExitCodeOp(errCode))
	cg.WriteOp(HaltOp)
}

// Finish add the final successful exit, and re-integrates with the parent gen.
func (alt *CodeGenFork) Finish() {
	alt.CodeGen.Finish()

	if len(alt.parent.buf) != alt.priorLen {
		panic("forked parent has assembled before copy finish")
	}
	if alt.contLabel != "" {
		alt.Label(alt.contLabel)
	}
	alt.contRef.WriteOffset(len(alt.buf))
	alt.parent.Assembler = alt.Assembler
}

// Finalize fills in the operation count limit, and returns a plan to run the
// compiled program.
func (cg *CodeGen) Finalize() word.Plan {
	// large upper limit for the search execution: run every operation for
	// every possible brute force solution
	cg.opLimRef.WriteValue1(uint16(fallFact(cg.Base, len(cg.Letters)) * int(cg.opCnt)))

	var err error
	cg.Plan.mach, err = NewTinyMachine(cg.Bytes(), cg.annos == nil)
	if err != nil {
		panic(err)
	}
	return &cg.Plan
}

// carryArg returns an argument containing the given column's carry.
func (cg *CodeGen) carryArg(col *word.Column) Arg {
	if col == nil {
		return Immediate(0)
	}

	if cr, defined := cg.regs.Get(cg.carryRegs[col.I]); defined {
		return Register(cr)
	}

	switch col.Carry {
	case word.CarryZero:
		fallthrough
	case word.CarryOne:
		return Immediate(uint16(col.Carry))
	}

	if rc := cg.computeCarry(col); rc != ArgNone {
		return rc
	}

	panic("unable to compute carry")
}

// setupCarryCompute sets up a register to compute the given column's carry.
func (cg *CodeGen) setupCarryCompute(col *word.Column) Arg {
	arg := cg.carryArg(col.Prior)
	if arg.Code.Register() == 0 {
		rc := Register(cg.regs.Take(cg.carryRegs[col.I]))
		cg.Annotatef("%s=%v", cg.carryRegs[col.I], arg.Val)
		cg.WriteOp(MoveOp(rc, arg))
		return rc
	}
	cg.regs.Reassign(cg.carryRegs[col.Prior.I], cg.carryRegs[col.I])
	return arg
}

// computeCarry adds code to compute a columns carry, or returns ArgNone if
// there isn't enough known letters to do so.
func (cg *CodeGen) computeCarry(col *word.Column) Arg {
	a, b := col.Chars[0], col.Chars[1]
	n := 0
	if a != 0 && !cg.Known[a] {
		n++
		return ArgNone
	}
	if b != 0 && !cg.Known[b] {
		n++
		return ArgNone
	}

	if cg.Annotated {
		cg.Label(fmt.Sprintf("computeCarry(%s)", col.Label()))
	}

	if n == 0 {
		// no letters to add means the columns value was just priorCarry (0/1)
		// which always has a carry out of 0.
		rc := cg.carryArg(col.Prior)
		if rc.Code.Register() == 0 {
			rc = Register(cg.regs.Take(cg.carryRegs[col.I]))
		} else {
			cg.regs.Reassign(cg.carryRegs[col.Prior.I], cg.carryRegs[col.I])
		}
		cg.Annotatef("%s=0", cg.carryRegs[col.I])
		cg.WriteOp(MoveOp(rc, Immediate(0)))
		return rc
	}

	rc := cg.setupCarryCompute(col)
	if a != 0 {
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rc, Location(cg.valOff[a])))
	}
	if b != 0 {
		// TODO: re-use any letter register
		cg.WriteOp(AddOp(rc, Location(cg.valOff[b])))
	}
	cg.WriteOp(DivOp(rc, Immediate(uint16(cg.Base))))
	return rc
}

// checkUsed adds code that exits with codeAlreadyUsed if the value in the
// given argument is already a used digit.
func (cg *CodeGen) checkUsed(rl Arg) (useLoc Arg) {
	ri := Register(cg.regs.Take("index"))
	useLoc = Indexed(cg.useOff[0], ri.Code.Register())
	cg.WriteOp(MoveOp(ri, rl))
	cg.WriteOp(MulOp(ri, Immediate(2)))
	cg.WriteOp(EQOp(useLoc, Immediate(0)))
	cg.WriteOp(JumpTOp(offExit))
	cg.WriteOp(setExitCodeOp(codeAlreadyUsed))
	cg.WriteOp(HaltOp)
	return
}

// checkAfterCompute adds code to check various column constraints after
// computation of one of its letters.
func (cg *CodeGen) checkAfterCompute(col *word.Column, c byte, rl, rc Arg) {
	if c == cg.Words[0][0] || c == cg.Words[1][0] || c == cg.Words[2][0] {
		cg.checkInitialLetter(col, c, rl)
	}
	cg.checkFixedCarry(col, rc)
}

// checkInitialLetter adds code to check that the given initial letter is
// non-zero.
func (cg *CodeGen) checkInitialLetter(col *word.Column, c byte, rl Arg) {
	if cg.Annotated {
		cg.Label(fmt.Sprintf("checkInitialLetter(%s)", string(c)))
	}
	cg.WriteOp(EQOp(rl, Immediate(0)))
	cg.WriteOp(JumpFOp(offExit))
	cg.WriteOp(setExitCodeOp(codeCheckFailed))
	cg.WriteOp(HaltOp)
}

// checkFixedCarry adds code to check that the computed carry matches the
// previously fixed value for this column (or generates no code if none was
// fixed.
func (cg *CodeGen) checkFixedCarry(col *word.Column, rc Arg) {
	switch col.Carry {
	case word.CarryZero:
		fallthrough
	case word.CarryOne:
		if cg.Annotated {
			cg.Label(fmt.Sprintf("checkFixedCarry(%s)", col.Label()))
		}
		cg.WriteOp(EQOp(rc, Immediate(uint16(col.Carry))))
		cg.WriteOp(JumpTOp(offExit))
		cg.WriteOp(setExitCodeOp(codeCheckFailed))
		cg.WriteOp(HaltOp)
	}
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
