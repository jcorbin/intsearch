package word

import (
	"fmt"
	"strings"
)

// CarryValue encodes knowledge about a carry out of a problem column being
// planned; possible values are unknown, eventually computed, or fixed (0/1).
type CarryValue int

const (
	// CarryUnknown means that the carry out of a column is unknown.
	CarryUnknown CarryValue = iota - 1

	// CarryZero means that the carry out of a column is fixed to 0.
	CarryZero

	// CarryOne means that the carry out of a column is fixed to 1.
	CarryOne

	// CarryComputed means that the carry out of a column will be computed (per
	// already generated plan).
	CarryComputed
)

//go:generate stringer -type=CarryValue

// Expr returns a single character describing the carry value
func (cv CarryValue) Expr() string {
	switch cv {
	case CarryUnknown:
		return "?"
	case CarryZero:
		return "0"
	case CarryOne:
		return "1"
	case CarryComputed:
		return "C"
	default:
		return "!"
	}
}

// Column describes a single column state in a problem under planning.
type Column struct {
	I       int
	Prior   *Column
	Chars   [3]byte
	Solved  bool
	Have    int
	Known   int
	Unknown int
	Fixed   int
	Carry   CarryValue
}

func (col *Column) String() string {
	return fmt.Sprintf(
		"%s solved=%t have=%d known=%d unknown=%d fixed=%d",
		col.Label(),
		col.Solved, col.Have, col.Known, col.Unknown, col.Fixed)
}

// Label returns a plan description of the column.
func (col *Column) Label() string {
	return fmt.Sprintf("[%d] %s carry=%s", col.I, col.expr(), col.Carry.Expr())
}

func (col *Column) expr() string {
	parts := make([]string, 0, 7)
	if col.Prior != nil {
		parts = append(parts, col.Prior.Carry.Expr())
	}
	for _, c := range col.Chars[:2] {
		if c != 0 {
			if len(parts) > 0 {
				parts = append(parts, "+", string(c))
			} else {
				parts = append(parts, string(c))
			}
		}
	}
	parts = append(parts, "=", string(col.Chars[2]))
	return strings.Join(parts, " ")
}
