package opcode

// Op is an instruction for a machine.  An op may have up to two arguments.
type Op struct {
	Code       OpCode
	Arg1, Arg2 Arg
}

// Arg is an operand for an Op.
type Arg struct {
	Code ArgCode
	Val  uint16
}

// ArgNone is the zero value argument that means "no argument".
var ArgNone = Arg{}

// Register builds an argument for a numbered register.
func Register(reg byte) Arg {
	reg = reg & argReg
	return Arg{ArgCode(reg), 0}
}

// Immediate builds an argument for an immediate value.
func Immediate(val uint16) Arg {
	return Arg{ArgCodeImm, val}
}

// Indirect builds an argument for a memory location contained in a register.
func Indirect(reg byte) Arg {
	reg = reg & argReg
	return Arg{ArgCodeInd | ArgCode(reg), 0}
}

// Location builds an argument for an immediate memory location.
func Location(addr uint16) Arg {
	return Arg{ArgCodeImm | ArgCodeInd, addr}
}

// Indexed builds an argument for an indexed memory location; an indexed
// location as an immediate base address plus the value of a register.
func Indexed(addr uint16, reg byte) Arg {
	reg = reg & argReg
	return Arg{ArgCodeImm | ArgCodeInd | ArgCode(reg), addr}
}

func (op Op) String() string {
	return op.Code.Format(
		op.Arg1.Code, op.Arg2.Code,
		op.Arg1.Val, op.Arg2.Val)
}

// HaltOp terminates the machine's execution loop; it's also incidentally the
// zero value Op.
var HaltOp = Op{} // == Op{HALT, ArgNone, ArgNone}

// OpLim set the operation count limit.
func OpLim(limit uint16) Op {
	return Op{OPLIM, Immediate(limit), ArgNone}
}

// MoveOp loads two bytes from memory into a register.
func MoveOp(dst, src Arg) Op {
	if dst.Code.Immediate() && !dst.Code.Indirect() {
		panic("invalid dst argument")
	}
	return Op{MOVE, dst, src}
}

// MoveLOp loads a byte from memory into the low byte of a register.
func MoveLOp(dst, src Arg) Op {
	if dst.Code.Immediate() && !dst.Code.Indirect() {
		panic("invalid dst argument")
	}
	return Op{MOVEL, dst, src}
}

// MoveHOp loads a byte from memory into the high byte of a register.
func MoveHOp(dst, src Arg) Op {
	if dst.Code.Immediate() && !dst.Code.Indirect() {
		panic("invalid dst argument")
	}
	return Op{MOVEH, dst, src}
}

// SwapOp exchanges two registers
func SwapOp(a, b byte) Op { return Op{SWAP, Register(a), Register(b)} }

// JumpOp increments PI by an immediate offset.
func JumpOp(offset int16) Op { return Op{JUMP, Immediate(uint16(offset)), ArgNone} }

// JumpFOp increments PI by an immediate offset, if the Cond is false.
func JumpFOp(offset int16) Op { return Op{JUMPF, Immediate(uint16(offset)), ArgNone} }

// JumpTOp increments PI by an immediate offset, if the Cond is true.
func JumpTOp(offset int16) Op { return Op{JUMPT, Immediate(uint16(offset)), ArgNone} }

// ForkOp defers a copy of the current machine state, and increments the copy's
// PI by an immediate offset.
func ForkOp(offset int16) Op { return Op{FORK, Immediate(uint16(offset)), ArgNone} }

// ForkFOp defers a copy of the current machine state, and increments the copy's
// PI by an immediate offset, only if Cond is false.
func ForkFOp(offset int16) Op { return Op{FORKF, Immediate(uint16(offset)), ArgNone} }

// ForkTOp defers a copy of the current machine state, and increments the copy's
// PI by an immediate offset, only if Cond is true.
func ForkTOp(offset int16) Op { return Op{FORKT, Immediate(uint16(offset)), ArgNone} }

// BranchOp defers a copy of the current machine state, and increments PI by an immediate offset.
func BranchOp(offset int16) Op { return Op{BRANCH, Immediate(uint16(offset)), ArgNone} }

// BranchFOp defers a copy of the current machine state, and increments PI by an immediate offset, only if Cond is false.
func BranchFOp(offset int16) Op { return Op{BRANCHF, Immediate(uint16(offset)), ArgNone} }

// BranchTOp defers a copy of the current machine state, and increments PI by an immediate offset, only if Cond is true.
func BranchTOp(offset int16) Op { return Op{BRANCHT, Immediate(uint16(offset)), ArgNone} }

// LTOp compares two registers, setting Cond = a < b.
func LTOp(a, b Arg) Op { return Op{LT, a, b} }

// LTEOp compares two registers, setting Cond = a <= b.
func LTEOp(a, b Arg) Op { return Op{LTE, a, b} }

// EQOp compares two registers, setting Cond = a == b.
func EQOp(a, b Arg) Op { return Op{EQ, a, b} }

// GTEOp compares two registers, setting Cond = a >= b.
func GTEOp(a, b Arg) Op { return Op{GTE, a, b} }

// GTOp compares two registers, setting Cond = a > b.
func GTOp(a, b Arg) Op { return Op{GT, a, b} }

// NegOp negates a value.
func NegOp(a Arg) Op {
	if a.Code.Immediate() && !a.Code.Indirect() {
		panic("invalid a argument")
	}
	return Op{NEG, a, ArgNone}
}

// SubOp decrements a value by another value.
func SubOp(a, b Arg) Op {
	if a.Code.Immediate() && !a.Code.Indirect() {
		panic("invalid a argument")
	}
	return Op{SUB, a, b}
}

// AddOp increments a value by another value.
func AddOp(a, b Arg) Op {
	if a.Code.Immediate() && !a.Code.Indirect() {
		panic("invalid a argument")
	}
	return Op{ADD, a, b}
}

// MulOp sets a value to its quotient by another value.
func MulOp(a, b Arg) Op {
	if a.Code.Immediate() && !a.Code.Indirect() {
		panic("invalid a argument")
	}
	return Op{MUL, a, b}
}

// DivOp sets a value to its quotient by another value.
func DivOp(a, b Arg) Op {
	if a.Code.Immediate() && !a.Code.Indirect() {
		panic("invalid a argument")
	}
	return Op{DIV, a, b}
}

// ModOp sets a value to its remainder by another value.
func ModOp(a, b Arg) Op {
	if a.Code.Immediate() && !a.Code.Indirect() {
		panic("invalid a argument")
	}
	return Op{MOD, a, b}
}
