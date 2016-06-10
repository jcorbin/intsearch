package opcode

import (
	"fmt"
	"strings"
)

// OpCode is an operation byte code
type OpCode byte

const (
	// HALT terminates the machine's execution loop.
	HALT = OpCode(iota)

	// OPLIM sets the operation count limit.
	OPLIM // limit value

	// MOVE copies two bytes from src to dst.
	MOVE // dst reg|mem, src reg|mem|value

	// MOVEL copies the low byte from src to dst.
	MOVEL // dst reg|mem, src mem|mem|value

	// MOVEH copies the high byte from src to dst.
	MOVEH // dst reg|mem, src mem|mem|value

	// SWAP exchanges two registers
	SWAP // a, b reg

	// JUMP increments PI by an immediate offset.
	JUMP // target offset

	// JUMPF increments PI by an immediate offset, if the Cond is false.
	JUMPF // target offset

	// JUMPT increments PI by an immediate offset, if the Cond is true.
	JUMPT // target offset

	// FORK defers a copy of the current machine state, and increments the
	// copy's PI by an immediate offset.
	FORK // child offset

	// FORKF defers a copy of the current machine state, and increments the
	// copy's PI by an immediate offset, only if Cond is false.
	FORKF // child offset

	// FORKT defers a copy of the current machine state, and increments the
	// copy's PI by an immediate offset, only if Cond is true.
	FORKT // child offset

	// BRANCH defers a copy of the current machine state, and increments PI by
	// an immediate offset.
	BRANCH // parent offset

	// BRANCHF defers a copy of the current machine state, and increments PI by
	// an immediate offset, only if Cond is false.
	BRANCHF // parent offset

	// BRANCHT defers a copy of the current machine state, and increments PI by
	// an immediate offset, only if Cond is true.
	BRANCHT // parent offset

	// LT compares two registers and/or immediate values, setting Cond = a < b.
	LT // a, b reg|value

	// LTE compares two registers and/or immediate values, setting Cond = a <= b.
	LTE // a, b reg|value

	// EQ compares two registers and/or immediate values, setting Cond = a == b.
	EQ // a, b reg|value

	// GTE compares two registers and/or immediate values, setting Cond = a >= b.
	GTE // a, b reg|value

	// GT compares two registers and/or immediate values, setting Cond = a > b.
	GT // a, b reg|value

	// NEG negates the value of a register.
	NEG // a reg

	// SUB decrements a register by another register, or an immediate value.
	SUB // a reg, b reg|value

	// ADD increments a register by another register, or an immediate value.
	ADD // a reg, b reg|value

	// MUL multiplies a register by another register, or an immediate value.
	MUL // a reg, b reg|value

	// DIV sets a register to its quotient by another register, or an immediate value.
	DIV // a reg, b reg|value

	// MOD sets a register to its remainder by another register, or an immediate value.
	MOD // a reg, b reg|value
)

var opMeta = []struct {
	arity  int
	name   string
	immed1 string
	immed2 string
}{
	{0, "HALT", "INVALID", "INVALID"},    // HALT
	{1, "OPLIM", "$value", "INVALID"},    // OPLIM
	{2, "MOVE", "INVALID", "$value"},     // MOVE
	{2, "MOVEL", "INVALID", "$value"},    // MOVEL
	{2, "MOVEH", "INVALID", "$value"},    // MOVEH
	{2, "SWAP", "INVALID", "INVALID"},    // SWAP
	{1, "JUMP", "+offset", "INVALID"},    // JUMP
	{1, "JUMPF", "+offset", "INVALID"},   // JUMPF
	{1, "JUMPT", "+offset", "INVALID"},   // JUMPT
	{1, "FORK", "+offset", "INVALID"},    // FORK
	{1, "FORKF", "+offset", "INVALID"},   // FORKF
	{1, "FORKT", "+offset", "INVALID"},   // FORKT
	{1, "BRANCH", "+offset", "INVALID"},  // BRANCH
	{1, "BRANCHF", "+offset", "INVALID"}, // BRANCHF
	{1, "BRANCHT", "+offset", "INVALID"}, // BRANCHT
	{2, "LT", "$value", "$value"},        // LT
	{2, "LTE", "$value", "$value"},       // LTE
	{2, "EQ", "$value", "$value"},        // EQ
	{2, "GTE", "$value", "$value"},       // GTE
	{2, "GT", "$value", "$value"},        // GT
	{1, "NEG", "INVALID", "INVALID"},     // NEG
	{2, "SUB", "INVALID", "$value"},      // SUB
	{2, "ADD", "INVALID", "$value"},      // ADD
	{2, "MUL", "INVALID", "$value"},      // MUL
	{2, "DIV", "INVALID", "$value"},      // DIV
	{2, "MOD", "INVALID", "$value"},      // MOD
}

var opArgImmedFmt = map[string]func(uint16) string{
	"$value":   func(val uint16) string { return fmt.Sprintf("%d", val) },
	"+offset":  func(val uint16) string { return fmt.Sprintf("%+05x", int16(val)) },
	"@address": func(val uint16) string { return fmt.Sprintf("@%04x", val) },
}

// Name returns the operation name.
func (code OpCode) Name() string {
	if int(code) < len(opMeta) {
		return opMeta[code].name
	}
	return "INVALID"
}

// Arity returns how many arguments the operation may have.
func (code OpCode) Arity() int {
	return opMeta[code].arity
}

// String returns a human friendly string describing the operation and its
// argument types.
func (code OpCode) String() string {
	if int(code) < len(opMeta) {
		return opMeta[code].name
	}
	return fmt.Sprintf("INVALID(%02x)", byte(code))
}

// ArgCode is an operand byte code
type ArgCode byte

const (
	// ArgCodeNone codes for no argument value.
	ArgCodeNone = ArgCode(0x00)

	// ArgCodeImm codes an immediate argument value.
	ArgCodeImm = 0x80

	// ArgCodeInd codes an indirect argument.
	ArgCodeInd = 0x40

	argFlags = 0xc0
	argReg   = 0x3f
)

// Immediate returns true if the argument has an immediate component.
func (code ArgCode) Immediate() bool {
	return code&ArgCodeImm == ArgCodeImm
}

// Indirect returns true if the argument is an indirect memory location.
func (code ArgCode) Indirect() bool {
	return code&ArgCodeInd == ArgCodeInd
}

// Register returns the argument register number.
func (code ArgCode) Register() byte {
	return byte(code & argReg)
}

func (code ArgCode) String() string {
	if code&ArgCodeInd != 0 {
		if code&ArgCodeImm != 0 {
			return fmt.Sprintf("@($IMMED + %%%d)", byte(code&argReg))
		}
		return fmt.Sprintf("@%%%d", byte(code&argReg))
	}
	if code&ArgCodeImm != 0 {
		return "$IMMED"
	}
	return fmt.Sprintf("%%%d", byte(code&argReg))
}

// Format returns an argument string specialized to a particular immediate
// value and role (if needed).
func (code ArgCode) Format(val uint16, role string) string {
	if code == ArgCodeNone {
		if role != "INVALID" {
			return fmt.Sprintf("MISSING(%s)", role)
		}
		return "MISSING"
	}
	if code&ArgCodeInd != 0 {
		if code&ArgCodeImm != 0 {
			if code&argReg != 0 {
				return fmt.Sprintf("@(%04x + %%%d)", val, byte(code&argReg))
			}
			return fmt.Sprintf("@%04x", val)
		}
		if code&argReg != 0 {
			return fmt.Sprintf("@%%%d", code&argReg)
		}
		return "@¡0000!"
	}
	if code&ArgCodeImm != 0 {
		if f := opArgImmedFmt[role]; f != nil {
			return f(val)
		}
		return fmt.Sprintf("UNKNOWN(%q, %d)", role, val)
	}
	return fmt.Sprintf("%%%d", byte(code&argReg))
}

// Format returns an op string specialized to two particular arg codes and
// immediate values (if needed).
func (code OpCode) Format(a1, a2 ArgCode, v1, v2 uint16) string {
	if int(code) >= len(opMeta) {
		return fmt.Sprintf("INVALID(%02x, %02x, %02x)", byte(code), byte(a1), byte(a2))
	}

	var parts [3]string
	n := opMeta[code].arity

	k := 1
	parts[0] = opMeta[code].name
	if n > 0 {
		parts[1] = a1.Format(v1, opMeta[code].immed1)
		k = 2
		if n > 1 {
			parts[2] = a2.Format(v2, opMeta[code].immed2)
			k = 3
		}
	}

	if n < 1 && a1 != ArgCodeNone {
		parts[1] = fmt.Sprintf("EXTRA(%02x, %d)", byte(a1), v1)
		k = 2
	}
	if n < 2 && a2 != ArgCodeNone {
		parts[2] = fmt.Sprintf("EXTRA(%02x, %d)", byte(a2), v2)
		k = 3
	}

	return strings.Join(parts[:k], " ")
}
