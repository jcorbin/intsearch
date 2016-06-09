package opcode

import (
	"encoding/binary"
	"fmt"
)

// Assembler builds a byte buffer of encoded ops
type Assembler struct {
	buf   []byte
	bo    binary.ByteOrder
	opCnt uint
}

// Ref is a reference to an assembled operation so that its immediate args can
// be filled in later.
type Ref struct {
	as *Assembler
	i  int
	j  int
	k  int
	l  int
}

func (ref Ref) String() string {
	op, _ := DecodeOp(ref.as.bo, ref.as.buf, ref.i)
	return fmt.Sprintf("@%04x %v", ref.i, op)
}

// Offset returns the offset of the referenced operation.
func (ref Ref) Offset() int {
	return ref.i
}

// EndOffset returns the offset after the referenced operation.
func (ref Ref) EndOffset() int {
	if ref.k != 0 {
		return ref.k + 2
	}
	if ref.j != 0 {
		return ref.j + 2
	}
	panic("invalid Ref, no immediate arg!")
}

// Arg1Offset returns the offset of the referenced operation's immedeiate arg1
// value, or zero if the operation has none.
func (ref Ref) Arg1Offset() int {
	return ref.j
}

// Arg2Offset returns the offset of the referenced operation's immedeiate arg1
// value, or zero if the operation has none.
func (ref Ref) Arg2Offset() int {
	return ref.k
}

// WriteValue1 fills in the immediate arg1 value.
func (ref Ref) WriteValue1(val uint16) {
	if ref.j == 0 {
		panic("arg1 isn't immediate")
	}
	ref.as.bo.PutUint16(ref.as.buf[ref.j:], val)
}

// WriteValue2 fills in the immediate arg1 value.
func (ref Ref) WriteValue2(val uint16) {
	if ref.k == 0 {
		panic("arg1 isn't immediate")
	}
	ref.as.bo.PutUint16(ref.as.buf[ref.k:], val)
}

// WriteOffset writes value1 as the difference between the given offset and
// the offset after the referenced operation.
func (ref Ref) WriteOffset(offset int) {
	ref.WriteValue1(uint16(offset - ref.l))
}

// NewAssembler creates a new assembler with the default capacity (4k)
func NewAssembler(bo binary.ByteOrder) *Assembler {
	return NewAssemblerSize(bo, 4096)
}

// NewAssemblerSize creates a new assembler with the given buffer size.
func NewAssemblerSize(bo binary.ByteOrder, n int) *Assembler {
	return &Assembler{
		bo:  bo,
		buf: make([]byte, 0, n),
	}
}

// Dump prints the assembled program.
func (as *Assembler) Dump(logf func(format string, args ...interface{})) {
	for i := 0; i < len(as.buf); {
		op, j := DecodeOp(as.bo, as.buf, i)
		logf("%04x: %v", i, op)
		i = j
	}
}

// Bytes returns the assembled byte slice.
func (as *Assembler) Bytes() []byte {
	return as.buf
}

// WriteOp encodes an operation into the internal byte buffer.
func (as *Assembler) WriteOp(op Op) int {
	i := as.grow(7)
	i = op.Encode(as.bo, as.buf, i)
	as.buf = as.buf[:i]
	as.opCnt++
	return i
}

// WriteOpRef encodes an operation into the internal byte buffer, and returns a
// Ref to it.
func (as *Assembler) WriteOpRef(op Op) Ref {
	i := as.grow(7)
	j := op.Encode(as.bo, as.buf, i)
	as.buf = as.buf[:j]
	if op.Arg2.Code.Immediate() && op.Arg1.Code.Immediate() {
		return Ref{as: as, i: i, j: j - 4, k: j - 2, l: j}
	} else if op.Arg1.Code.Immediate() {
		return Ref{as: as, i: i, j: j - 2, k: 0, l: j}
	} else if op.Arg2.Code.Immediate() {
		return Ref{as: as, i: i, j: 0, k: j - 2, l: j}
	}
	panic("invalid WriteOpRef op: no immediate arg(s)")
}

func (as *Assembler) grow(n int) int {
	i := len(as.buf)
	j := i + int(n)
	if j > cap(as.buf) {
		k := 2*cap(as.buf) + int(n)
		if k > 0xffff {
			k = 0xffff
		}
		if j > k {
			panic("assembler overflow")
		}
		buf := make([]byte, k)
		copy(buf, as.buf)
		as.buf = buf
	}
	as.buf = as.buf[:j]
	return i
}
