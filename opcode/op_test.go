package opcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcorbin/intsearch/opcode"
)

var r1 = opcode.Register(1)

var stringTestCases = []struct {
	expected []string
	ops      []opcode.Op
}{
	{
		expected: []string{
			"JUMP +0000",
			"JUMPT -0001",
			"MOVE %1 1",
			"SUB %1 2",
			"NEG %1",
			"MOVE @002a %1",
			"HALT",
			"MOVE @0020 MISSING($value)",
			"MOVE MISSING %5",
		},
		ops: []opcode.Op{
			opcode.JumpOp(0),
			opcode.JumpTOp(-1),
			opcode.MoveOp(r1, opcode.Immediate(1)),
			opcode.SubOp(r1, opcode.Immediate(2)),
			opcode.NegOp(r1),
			opcode.MoveOp(opcode.Location(42), r1),
			opcode.HaltOp,
			opcode.MoveOp(opcode.Location(32), opcode.ArgNone),
			opcode.MoveOp(opcode.ArgNone, opcode.Register(5)),
		},
	},
}

func TestOp_String(t *testing.T) {
	for _, c := range stringTestCases {
		for i, op := range c.ops {
			if got := op.String(); got != c.expected[i] {
				assert.Equal(t, c.expected[i], got, "unexpected output for op[%d] %s", i, op.Code.Name())
			}
		}
	}
}
