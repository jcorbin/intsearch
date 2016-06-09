package opcode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcorbin/intsearch/opcode"
)

func TestTinyMachine_RunAll(t *testing.T) {
	as := opcode.NewAssembler(opcode.TinyMachineByteOrder())
	as.WriteOp(opcode.MoveOp(r1, opcode.Immediate(0)))
	forkRef := as.WriteOpRef(opcode.BranchOp(0))      // :loop
	as.WriteOp(opcode.AddOp(r1, opcode.Immediate(1))) // :next
	as.WriteOp(opcode.LTOp(r1, opcode.Immediate(9)))
	jumpRef := as.WriteOpRef(opcode.JumpTOp(0))
	contRef := as.WriteOpRef(opcode.MoveLOp(opcode.Location(0x0001), r1)) // :cont
	as.WriteOp(opcode.HaltOp)

	forkRef.WriteOffset(contRef.Offset())
	jumpRef.WriteOffset(forkRef.Offset())

	mach, err := opcode.NewTinyMachine(as.Bytes(), false)
	require.NoError(t, err)

	res := machRes{
		T:  t,
		bs: make([]byte, 0, 10),
	}
	final := mach.RunAll(&res)
	if final != nil {
		t.Logf("machine check error: %v", final.Check())
	}
	require.Nil(t, final, "machine check error")
	assert.Equal(t, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, res.bs)

	if t.Failed() {
		mach.Reset()
		require.Nil(t, mach.RunAll(opcode.NewTracePrinter()), "machine check error")
	}
}

type machRes struct {
	*testing.T
	bs       []byte
	state    interface{}
	beforeOp opcode.Op
	afterOp  opcode.Op
}

func (res *machRes) collect(mach opcode.Machine) [1]byte {
	var buf [1]byte
	require.Equal(res.T, 1, mach.CopyMemory(1, buf[:]))
	res.bs = append(res.bs, buf[0])
	return buf
}

func (res *machRes) Result(mach opcode.Machine) bool {
	res.collect(mach)
	return mach.Check() != nil
}

func (res *machRes) Before(mach opcode.Machine) {
	res.state = mach.State()
	res.beforeOp = mach.NextOp()
}

func (res *machRes) After(mach opcode.Machine) {
	res.afterOp = mach.LastOp()
	assert.Equal(res.T,
		res.beforeOp.String(),
		res.afterOp.String(),
		"saw op instability around %v", res.state)
}

func (res *machRes) Emit(action string, parent, child opcode.Machine) {
}
