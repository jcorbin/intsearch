package opcode

import "fmt"

type opError struct {
	op   Op
	desc string
}

func (oe opError) Error() string {
	return fmt.Sprintf("invalid op %v: %s", oe.op, oe.desc)
}

func opErrorf(op Op, format string, args ...interface{}) opError {
	if len(args) == 0 {
		return opError{op, format}
	}
	return opError{op, fmt.Sprintf(format, args...)}
}

// Validate returns a specific error if the operation is invalid.  Checks basic
// things like valid op code, and valid arg codes for the given op code.
func (op Op) Validate() error {
	switch op.Code {
	case HALT: // --
		if op.Arg1.Code != 0 || op.Arg2.Code != 0 {
			return opErrorf(op, "expected no arguments")
		}

	case OPLIM: // limit value
		if op.Arg1.Code.Indirect() || !op.Arg1.Code.Immediate() {
			return opErrorf(op, "expected immediate arg1")
		}
		if op.Arg2.Code != 0 {
			return opErrorf(op, "expected no arg2")
		}

	case MOVE: // dst reg|mem, src val
		fallthrough
	case MOVEL: // dst reg|mem, src val
		fallthrough
	case MOVEH: // dst reg|mem, src val
		if op.Arg1.Code.Immediate() && !op.Arg1.Code.Indirect() {
			return opErrorf(op, "expected arg1 to be register or memory")
		}
		if op.Arg2.Code == 0 {
			return opErrorf(op, "missing arg2")
		}

	case SWAP: // a, b reg|mem
		if op.Arg1.Code.Immediate() && !op.Arg1.Code.Indirect() {
			return opErrorf(op, "expected arg1 to be register or memory")
		}
		if op.Arg2.Code.Immediate() && !op.Arg2.Code.Indirect() {
			return opErrorf(op, "expected arg2 to be register or memory")
		}

	case JUMP: // offset imm
		fallthrough
	case JUMPF: // offset imm
		fallthrough
	case JUMPT: // offset imm
		fallthrough
	case FORK: // offset imm
		fallthrough
	case FORKF: // offset imm
		fallthrough
	case FORKT: // offset imm
		fallthrough
	case BRANCH: // offset imm
		fallthrough
	case BRANCHF: // offset imm
		fallthrough
	case BRANCHT: // offset imm
		if !op.Arg1.Code.Immediate() || op.Arg1.Code.Indirect() {
			return opErrorf(op, "expected arg1 offset")
		}
		if op.Arg2.Code != 0 {
			return opErrorf(op, "expected no arg2")
		}

	case LT: // a, b reg|mem|imm
		fallthrough
	case LTE: // a, b reg|mem|imm
		fallthrough
	case EQ: // a, b reg|mem|imm
		fallthrough
	case GTE: // a, b reg|mem|imm
		fallthrough
	case GT: // a, b reg|mem|imm
		if op.Arg1.Code == 0 {
			return opErrorf(op, "missing arg1")
		}
		if op.Arg2.Code == 0 {
			return opErrorf(op, "missing arg2")
		}

	case NEG: // dst reg|mem
		if op.Arg1.Code.Immediate() && !op.Arg1.Code.Indirect() {
			return opErrorf(op, "expected arg1 to be register or memory")
		}
		if op.Arg2.Code != 0 {
			return opErrorf(op, "expected no arg2")
		}

	case SUB: // dst reg|mem, val reg|mem|imm
		fallthrough
	case ADD: // dst reg|mem, val reg|mem|imm
		fallthrough
	case MOD: // dst reg|mem, val reg|mem|imm
		fallthrough
	case MUL: // dst reg|mem, val reg|mem|imm
		fallthrough
	case DIV: // dst reg|mem, val reg|mem|imm
		if op.Arg1.Code.Immediate() && !op.Arg1.Code.Indirect() {
			return opErrorf(op, "expected arg1 to be register or memory")
		}
		if op.Arg2.Code == 0 {
			return opErrorf(op, "missing arg2")
		}

	default:
		return opErrorf(op, "invalid opcode")
	}
	return nil
}
