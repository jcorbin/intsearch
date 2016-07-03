package opcode

import (
	"encoding/binary"
	"flag"
	"fmt"
	"sync"
	"unsafe"
)

var debug = flag.Bool("tinymachine.debug", false, "debug tinymachine running")

const (
	tinyMachineMemorySize   = 256
	tinyMachineNumRegisters = 5
)

var (
	bigOne      = [2]byte{1, 0}
	isBigEndian = *(*uint16)(unsafe.Pointer(&bigOne[0])) != 1
)

// TinyMachineByteOrder returns the binary byte order used by TinyMachine.
func TinyMachineByteOrder() binary.ByteOrder {
	if isBigEndian {
		return binary.BigEndian
	}
	return binary.LittleEndian
}

// TinyMachine is a 3-register machine with 64 bytes of memory.
type TinyMachine struct {
	bo  binary.ByteOrder
	p   []byte
	pi  uint16
	op  Op
	h   bool
	t   bool
	r1  uint16
	r2  uint16
	r3  uint16
	r4  uint16
	r5  uint16
	ol  uint16
	oc  uint16
	m   [tinyMachineMemorySize]byte
	ctx *tinyMachineCtx
}

// NewTinyMachine creates a new tiny machine with a loaded program, returning
// any load error.
func NewTinyMachine(p []byte, assumeValid bool) (Machine, error) {
	mach := &TinyMachine{
		bo: TinyMachineByteOrder(),
	}
	return mach, mach.Load(p, assumeValid)
}

type tinyMachineCtx struct {
	sync.Pool
	root     *TinyMachine
	rootFree bool
	result   Resultor
	trace    Tracer
	off      int
	q        []*TinyMachine
}

func (ctx *tinyMachineCtx) init(res Resultor) {
	if ctx.off != 0 {
		ctx.prune()
	}
	if len(ctx.q) != 0 {
		panic("unimplemented")
	}
	trc, _ := res.(Tracer)
	ctx.result = res
	ctx.trace = trc
	ctx.rootFree = false
}

func (ctx *tinyMachineCtx) fork(parent *TinyMachine, offset uint16) {
	child := parent.Copy()
	child.pi += offset
	if ctx.trace != nil {
		ctx.trace.Emit("FORK", parent, child)
	}
	ctx.pushq(child)
}

func (ctx *tinyMachineCtx) branch(parent *TinyMachine, offset uint16) {
	child := parent.Copy()
	parent.pi += offset
	if ctx.trace != nil {
		ctx.trace.Emit("BRANCH", parent, child)
	}
	ctx.pushq(child)
}

func (ctx *tinyMachineCtx) Put(other *TinyMachine) {
	if other == ctx.root {
		ctx.rootFree = true
		return
	}
	ctx.Pool.Put(other)
}

func (ctx *tinyMachineCtx) Get() *TinyMachine {
	if ctx.rootFree {
		ctx.rootFree = false
		return ctx.root
	}
	if item := ctx.Pool.Get(); item != nil {
		return item.(*TinyMachine)
	}
	return nil
}

func (ctx *tinyMachineCtx) shiftq() (other *TinyMachine) {
	if ctx.off < len(ctx.q) {
		other, ctx.q[ctx.off] = ctx.q[ctx.off], nil
		ctx.off++
		if *debug {
			fmt.Printf("%p SHIFT\n", other)
		}
	}
	return
}

func (ctx *tinyMachineCtx) pushq(other *TinyMachine) {
	if ctx.off > 0 {
		ctx.prune()
	}
	ctx.q = append(ctx.q, other)
	if *debug {
		fmt.Printf("%p PUSH\n", other)
	}
}

func (ctx *tinyMachineCtx) prune() {
	i := len(ctx.q) - ctx.off
	copy(ctx.q, ctx.q[ctx.off:])
	ctx.q = ctx.q[:i]
	ctx.off = 0
}

// Copy returns a copy of the current machine
func (mach *TinyMachine) Copy() *TinyMachine {
	other := mach.ctx.Get()
	if other == mach {
		panic("d'oh")
	}
	if other == nil {
		other = &TinyMachine{
			bo:  mach.bo,
			p:   mach.p,
			pi:  mach.pi,
			op:  mach.op,
			h:   mach.h,
			t:   mach.t,
			r1:  mach.r1,
			r2:  mach.r2,
			r3:  mach.r3,
			r4:  mach.r4,
			r5:  mach.r5,
			ol:  mach.ol,
			oc:  mach.oc,
			m:   mach.m,
			ctx: mach.ctx,
		}
		if *debug {
			fmt.Printf("%p ALLOC <- %p\n", other, mach)
		}
	} else {
		if *debug {
			fmt.Printf("%p REUSE <- %p\n", other, mach)
		}
		*other = *mach
	}
	return other
}

func (mach *TinyMachine) String() string {
	return fmt.Sprintf("TinyMachine(%v)", mach.State())
}

func (mach *TinyMachine) checkState() error {
	if int(mach.pi) > len(mach.p) {
		return ErrInvalidPI
	}
	if !mach.h {
		return ErrNotHalted
	}
	if mach.ol > 0 && mach.oc >= mach.ol {
		return ErrOpLimitExceeded{uint(mach.oc), uint(mach.ol)}
	}
	return nil
}

// Check checks the machine state for any error, and returns it.
func (mach *TinyMachine) Check() error {
	if err := mach.checkState(); err != nil {
		return err
	}
	if rv := mach.m[0]; rv != 0 {
		return ErrNonZeroHalt(rv)
	}
	return nil
}

// ByteOrder returns the correct byte ordering to read values out of the
// machine's memory.
func (mach *TinyMachine) ByteOrder() binary.ByteOrder {
	return mach.bo
}

// Reset resets the machine state to initial conditions.
func (mach *TinyMachine) Reset() {
	mach.pi = 0
	mach.h = false
	mach.t = false
	mach.r1 = 0
	mach.r2 = 0
	mach.r3 = 0
	mach.r4 = 0
	mach.r5 = 0
	mach.ol = 0
	mach.oc = 0
	for i := 0; i < tinyMachineMemorySize; i++ {
		mach.m[i] = 0
	}
}

// Program returns the loaded program
func (mach *TinyMachine) Program() []byte {
	return mach.p
}

// PI returns the current program index
func (mach *TinyMachine) PI() int {
	return int(mach.pi)
}

type tinyMachineState struct {
	PI uint16 `json:"pi"`
	H  bool   `json:"h"`
	T  bool   `json:"t"`
	R1 uint16 `json:"r1"`
	R2 uint16 `json:"r2"`
	R3 uint16 `json:"r3"`
	R4 uint16 `json:"r4"`
	R5 uint16 `json:"r5"`
	OC uint16 `json:"oc"`
	OL uint16 `json:"ol"`
}

func (st *tinyMachineState) String() string {
	if st.OL != 0 || st.OC != 0 {
		return fmt.Sprintf(
			"(%d/%d) @%04x h=%t t=%t r1=%d r2=%d r3=%d r4=%d r5=%d",
			st.OC, st.OL, st.PI, st.H, st.T, st.R1, st.R2, st.R3, st.R4, st.R5)
	}
	return fmt.Sprintf(
		"@%04x h=%t t=%t r1=%d r2=%d r3=%d r4=%d r5=%d",
		st.PI, st.H, st.T, st.R1, st.R2, st.R3, st.R4, st.R5)
}

// State returns an exported dump of all machine state.
func (mach *TinyMachine) State() interface{} {
	return &tinyMachineState{
		PI: mach.pi,
		H:  mach.h,
		T:  mach.t,
		R1: mach.r1,
		R2: mach.r2,
		R3: mach.r3,
		R4: mach.r4,
		R5: mach.r5,
		OC: mach.oc,
		OL: mach.ol,
	}
}

// CopyMemory copies memory from the machine at the given offset into the given
// buffer, returning the number of bytes copied.
func (mach *TinyMachine) CopyMemory(offset int, buf []byte) int {
	if offset < len(mach.m) {
		return copy(buf, mach.m[offset:])
	}
	return 0
}

// Load loads a program after validating it (unless told to assume validity).
func (mach *TinyMachine) Load(p []byte, assumeValid bool) error {
	if !assumeValid {
		if len(p) > 0xffff {
			return fmt.Errorf("program too long at %d bytes, maximum is %d", len(p), 0xffff)
		}
		for i := 0; i < len(p); {
			mach.op, i = DecodeOp(mach.bo, p, i)
			if err := mach.op.Validate(); err != nil {
				return err
			}
			if mach.op.Arg1.Code.Register() > tinyMachineNumRegisters {
				return opErrorf(mach.op, "invalid arg1 register, max is %d", tinyMachineNumRegisters)
			}
			if mach.op.Arg1.Code.Indirect() && mach.op.Arg1.Code.Immediate() && mach.op.Arg1.Val > tinyMachineMemorySize-2 {
				return opErrorf(mach.op, "invalid arg1 memory address, max is %d", tinyMachineMemorySize-2)
			}
			if mach.op.Arg2.Code.Register() > tinyMachineNumRegisters {
				return opErrorf(mach.op, "invalid arg2 register, max is %d", tinyMachineNumRegisters)
			}
			if mach.op.Arg2.Code.Indirect() && mach.op.Arg2.Code.Immediate() && mach.op.Arg2.Val > tinyMachineMemorySize-2 {
				return opErrorf(mach.op, "invalid arg2 memory address, max is %d", tinyMachineMemorySize-2)
			}
		}
	}
	mach.p = p
	mach.Reset()
	return nil
}

const (
	arg1Reg1 = 0x00000100
	arg1Reg2 = 0x00000200
	arg1Reg3 = 0x00000300
	arg2Reg1 = 0x00000001
	arg2Reg2 = 0x00000002
	arg2Reg3 = 0x00000003
)

// Run runs the, previously loaded, program on the machine.
func (mach *TinyMachine) Run() {
	if trc := mach.ctx.trace; trc != nil {
		for !mach.h && int(mach.pi) < len(mach.p) {
			trc.Before(mach)
			mach.Step()
			trc.After(mach)
		}
	} else {
		for !mach.h && int(mach.pi) < len(mach.p) {
			mach.Step()
		}
	}
}

// LastOp returns the op last decoded.
func (mach *TinyMachine) LastOp() Op {
	return mach.op
}

// NextOp decodes the next operation what will be ran; it does NOT advance pi.
func (mach *TinyMachine) NextOp() Op {
	nextOp, _ := DecodeOp(mach.bo, mach.p, int(mach.pi))
	return nextOp
}

func (mach *TinyMachine) resolveArgLoc(arg Arg) *uint16 {
	if arg.Code.Indirect() {
		return (*uint16)(unsafe.Pointer(&mach.m[mach.resolveAddr(arg)]))
	}
	if arg.Code.Immediate() {
		panic("cannot resolve location of immediate argument")
	}
	return mach.resolveRegisterLoc(arg)
}

func (mach *TinyMachine) resolveArgVal(arg Arg) uint16 {
	if arg.Code.Indirect() {
		return *((*uint16)(unsafe.Pointer(&mach.m[mach.resolveAddr(arg)])))
	}
	if arg.Code.Immediate() {
		if arg.Code.Register() != 0 {
			panic("invalid register number with an immediate flag")
		}
		return arg.Val
	}
	return mach.resolveRegisterVal(arg)
}

func (mach *TinyMachine) resolveAddr(arg Arg) (addr uint16) {
	if arg.Code.Immediate() {
		addr = arg.Val
	}
	switch arg.Code.Register() {
	case 0:
	case 0x01:
		addr += mach.r1
	case 0x02:
		addr += mach.r2
	case 0x03:
		addr += mach.r3
	case 0x04:
		addr += mach.r4
	case 0x05:
		addr += mach.r5
	default:
		panic("invalid register number")
	}
	return
}

func (mach *TinyMachine) resolveRegisterLoc(arg Arg) *uint16 {
	switch arg.Code.Register() {
	case 0:
		panic("null register")
	case 1:
		return &mach.r1
	case 2:
		return &mach.r2
	case 3:
		return &mach.r3
	case 4:
		return &mach.r4
	case 5:
		return &mach.r5
	}
	panic("invalid register number")
}

func (mach *TinyMachine) resolveRegisterVal(arg Arg) uint16 {
	switch arg.Code.Register() {
	case 0:
		panic("null register")
	case 1:
		return mach.r1
	case 2:
		return mach.r2
	case 3:
		return mach.r3
	case 4:
		return mach.r4
	case 5:
		return mach.r5
	}
	panic("invalid register number")
}

// Step executes a single machine operation.
func (mach *TinyMachine) Step() {
	if mach.h {
		return
	}

	i := int(mach.pi)
	mach.op, i = DecodeOp(mach.bo, mach.p, i)
	mach.pi = uint16(i)

	if mach.ol != 0 {
		mach.oc++
		if mach.oc >= mach.ol {
			mach.h = true
			return
		}
	}

	switch mach.op.Code {

	case OPLIM:
		val := mach.resolveArgVal(mach.op.Arg1)
		if mach.ol == 0 || val < mach.ol {
			mach.ol = val
		} else {
			panic("invalid attempt to raise op count limit")
		}

	case HALT:
		// halt is a single byte operation, rewind so that decode stays stuck.
		mach.h = true
		mach.pi--
		return

	case MOVE:
		loc := mach.resolveArgLoc(mach.op.Arg1)
		val := mach.resolveArgVal(mach.op.Arg2)
		*loc = val

	case MOVEL:
		loc := mach.resolveArgLoc(mach.op.Arg1)
		val := mach.resolveArgVal(mach.op.Arg2)
		*loc |= val & 0x00ff // TODO bit hacky

	case MOVEH:
		loc := mach.resolveArgLoc(mach.op.Arg1)
		val := mach.resolveArgVal(mach.op.Arg2)
		*loc |= val & 0xff00 // TODO bit hacky

	case SWAP: // a, b reg
		l1 := mach.resolveArgLoc(mach.op.Arg1)
		l2 := mach.resolveArgLoc(mach.op.Arg2)
		*l1, *l2 = *l2, *l1

	case JUMP: // offset
		off := mach.resolveArgVal(mach.op.Arg1)
		mach.pi += off
	case JUMPF: // offset
		if !mach.t {
			off := mach.resolveArgVal(mach.op.Arg1)
			mach.pi += off
		}
	case JUMPT: // offset
		if mach.t {
			off := mach.resolveArgVal(mach.op.Arg1)
			mach.pi += off
		}

	case FORK: // offset
		off := mach.resolveArgVal(mach.op.Arg1)
		mach.ctx.fork(mach, off)
	case FORKF: // offset
		if !mach.t {
			off := mach.resolveArgVal(mach.op.Arg1)
			mach.ctx.fork(mach, off)
		}
	case FORKT: // offset
		if mach.t {
			off := mach.resolveArgVal(mach.op.Arg1)
			mach.ctx.fork(mach, off)
		}

	case BRANCH: // offset
		off := mach.resolveArgVal(mach.op.Arg1)
		mach.ctx.branch(mach, off)
	case BRANCHF: // offset
		if !mach.t {
			off := mach.resolveArgVal(mach.op.Arg1)
			mach.ctx.branch(mach, off)
		}
	case BRANCHT: // offset
		if mach.t {
			off := mach.resolveArgVal(mach.op.Arg1)
			mach.ctx.branch(mach, off)
		}

	case LT: // a, b val
		a := mach.resolveArgVal(mach.op.Arg1)
		b := mach.resolveArgVal(mach.op.Arg2)
		mach.t = int16(a) < int16(b)
	case LTE: // a, b val
		a := mach.resolveArgVal(mach.op.Arg1)
		b := mach.resolveArgVal(mach.op.Arg2)
		mach.t = int16(a) <= int16(b)
	case EQ: // a, b val
		a := mach.resolveArgVal(mach.op.Arg1)
		b := mach.resolveArgVal(mach.op.Arg2)
		mach.t = a == b
	case GTE: // a, b val
		a := mach.resolveArgVal(mach.op.Arg1)
		b := mach.resolveArgVal(mach.op.Arg2)
		mach.t = int16(a) >= int16(b)
	case GT: // a, b val
		a := mach.resolveArgVal(mach.op.Arg1)
		b := mach.resolveArgVal(mach.op.Arg2)
		mach.t = int16(a) > int16(b)

	case NEG: // dst val
		loc := mach.resolveArgLoc(mach.op.Arg1)
		*loc = -*loc
	case SUB: // dst reg|mem, val val
		loc := mach.resolveArgLoc(mach.op.Arg1)
		val := mach.resolveArgVal(mach.op.Arg2)
		*loc -= val
	case ADD: // dst reg|mem, val imm
		loc := mach.resolveArgLoc(mach.op.Arg1)
		val := mach.resolveArgVal(mach.op.Arg2)
		*loc += val
	case MUL: // dst reg|mem, val imm
		loc := mach.resolveArgLoc(mach.op.Arg1)
		v2 := mach.resolveArgVal(mach.op.Arg2)
		*loc = uint16(int16(*loc) * int16(v2))
	case DIV: // dst reg|mem, val imm
		loc := mach.resolveArgLoc(mach.op.Arg1)
		v2 := mach.resolveArgVal(mach.op.Arg2)
		*loc = uint16(int16(*loc) / int16(v2))
	case MOD: // dst reg|mem, val imm
		loc := mach.resolveArgLoc(mach.op.Arg1)
		v2 := mach.resolveArgVal(mach.op.Arg2)
		*loc = uint16(int16(*loc) % int16(v2))

	default:
		panic("invalid mach.op")
	}
	return
}

// RunAll runs the machine, and continues to run all deferred copies until none
// are left, or until the result function says to stop.
//
// Each finished machine is passed to the result function.  If the result
// function returns true, then the run loop stops and that machine is returned.
//
// Therefore, there are two choices for result function implementors:
// - if stop-on-first-result behavior is desired, implement a simple
//   predicate function, the machine returned by RunAll can then be used to
//   extract result data.
// - if multiple results are possible, or if you don't want to just take the
//   first, implement a result function that extracts data from each finished
//   machine and stores it somehow.
//
// Result functions MUST NOT retain references to the passed machines, as they
// may mutate after the result function returns.
func (mach *TinyMachine) RunAll(res Resultor) Machine {
	if mach.ctx != nil {
		if *debug {
			fmt.Printf("\n%p RUNALL ctx reset\n", mach)
		}
		mach.Reset()
	} else {
		if *debug {
			fmt.Printf("\n%p RUNALL new ctx\n", mach)
		}
		mach.ctx = &tinyMachineCtx{root: mach}
	}

	ctx := mach.ctx
	ctx.init(res)

	i := 0
	for ; mach != nil; mach = ctx.shiftq() {
		if *debug {
			fmt.Printf("%p RUN %v\n", mach, mach)
		}
		mach.Run()
		if mach.Check() == nil {
			i++
			if *debug {
				fmt.Printf("%p SOL_%d %v\n", mach, i, mach)
			}
		}
		if ctx.result.Result(mach) {
			return mach
		}
		if err := mach.checkState(); err != nil {
			return mach
		}
		if *debug {
			fmt.Printf("%p PUT\n", mach)
		}
		ctx.Put(mach)
	}
	return nil
}
