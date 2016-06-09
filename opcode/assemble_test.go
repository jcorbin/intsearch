package opcode_test

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcorbin/intsearch/opcode"
)

func TestAssembler(t *testing.T) {
	as := opcode.NewAssembler(binary.BigEndian)
	as.WriteOp(opcode.MoveOp(r1, opcode.Immediate(0)))
	forkRef := as.WriteOpRef(opcode.ForkOp(0))        // :loop
	as.WriteOp(opcode.AddOp(r1, opcode.Immediate(1))) // :next
	as.WriteOp(opcode.LTOp(r1, opcode.Immediate(9)))
	jumpRef := as.WriteOpRef(opcode.JumpTOp(0))
	contRef := as.WriteOpRef(opcode.MoveLOp(opcode.Location(0x0001), r1)) // :cont
	as.WriteOp(opcode.HaltOp)

	forkRef.WriteOffset(contRef.Offset())
	jumpRef.WriteOffset(forkRef.Offset())

	assert.Equal(t, []byte{
		// 0x00
		byte(opcode.MOVE), 0x01, 0x80, 0x00, 0x00,
		// 0x05
		byte(opcode.FORK), 0x80, 0x00, 0x0e,
		// 0x09
		byte(opcode.ADD), 0x01, 0x80, 0x00, 0x01,
		// 0x0e
		byte(opcode.LT), 0x01, 0x80, 0x00, 0x09,
		// 0x13
		byte(opcode.JUMPT), 0x80, 0xff, 0xee,
		// 0x17
		byte(opcode.MOVEL), 0xc0, 0x01, 0x00, 0x01,
		// 0x1c
		byte(opcode.HALT),
		// 0x1d
	}, as.Bytes())
}
