package word

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
