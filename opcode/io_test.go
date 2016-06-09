package opcode_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jcorbin/intsearch/opcode"
)

var ioTestCases = []struct {
	bytes []byte
	ops   []opcode.Op
}{
	{[]byte{
		byte(opcode.JUMP), 0x80, 0x00, 0x00,
		byte(opcode.JUMPT), 0x80, 0xff, 0xff,
		byte(opcode.MOVE), 0x01, 0x80, 0x00, 0x01,
		byte(opcode.SUB), 0x01, 0x80, 0x00, 0x02,
		byte(opcode.NEG), 0x01,
		byte(opcode.MOVE), 0xc0, 0x01, 0x00, 0x2a,
		byte(opcode.HALT),
	}, []opcode.Op{
		opcode.JumpOp(0),
		opcode.JumpTOp(-1),
		opcode.MoveOp(r1, opcode.Immediate(1)),
		opcode.SubOp(r1, opcode.Immediate(2)),
		opcode.NegOp(r1),
		opcode.MoveOp(opcode.Location(42), r1),
		opcode.HaltOp,
	}},
}

func TestOp_Encode(t *testing.T) {
	for _, c := range ioTestCases {
		var buf bytes.Buffer
		for _, op := range c.ops {
			var p [7]byte
			j := op.Encode(binary.BigEndian, p[:], 0)
			buf.Write(p[:j])
		}
		assert.Equal(t, c.bytes, buf.Bytes())
	}
}

func TestDecodeOp(t *testing.T) {
	for _, c := range ioTestCases {
		var ops []opcode.Op
		for i := 0; i < len(c.bytes); {
			var op opcode.Op
			op, i = opcode.DecodeOp(binary.BigEndian, c.bytes, i)
			ops = append(ops, op)
		}
		assert.Equal(t, c.ops, ops)
	}
}

func TestOp_EncodedSize(t *testing.T) {
	for _, testCase := range []struct {
		op   opcode.Op
		size int
		bs   []byte
	}{
		{opcode.JumpOp(8),
			2 + 2, []byte{
				byte(opcode.JUMP),
				opcode.ArgCodeImm,
				0x00, 0x08}},
		{opcode.MoveLOp(opcode.Location(0x0000), opcode.Immediate(42)),
			3 + 2 + 2, []byte{
				byte(opcode.MOVEL),
				opcode.ArgCodeImm | opcode.ArgCodeInd,
				opcode.ArgCodeImm,
				0x00, 0x00,
				0x00, 0x2a}},
		{opcode.HaltOp, 1, []byte{byte(opcode.HALT)}},
	} {
		var buf [10]byte
		i := testCase.op.Encode(binary.BigEndian, buf[:], 0)
		assert.Equal(t, testCase.size, testCase.op.EncodedSize(), "expected encoded size for %v", testCase.op)
		assert.Equal(t, testCase.size, i, "expected encode offset for %v", testCase.op)
		assert.Equal(t, testCase.bs, buf[:i], "expected encode output for %v", testCase.op)
	}
}
