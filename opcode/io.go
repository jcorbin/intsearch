package opcode

import "encoding/binary"

// EncodedSize returns how many bytes the op will encode to.
func (op Op) EncodedSize() (s int) {
	s++
	if n := op.Code.Arity(); n > 0 {
		s++
		if op.Arg1.Code.Immediate() {
			s += 2
		}
		if n > 1 {
			s++
			if op.Arg2.Code.Immediate() {
				s += 2
			}
		}
	}
	return
}

// Encode encodes Op into buf at i, returning the index after the newly encoded
// op.  Encode panics if the buffer isn't large enough.
func (op Op) Encode(bo binary.ByteOrder, buf []byte, i int) int {
	buf[i] = byte(op.Code)
	i++
	if n := op.Code.Arity(); n > 0 {
		buf[i] = byte(op.Arg1.Code)
		i++
		if n > 1 {
			buf[i] = byte(op.Arg2.Code)
			i++
			if op.Arg1.Code.Immediate() {
				bo.PutUint16(buf[i:], op.Arg1.Val)
				i += 2
			}
			if op.Arg2.Code.Immediate() {
				bo.PutUint16(buf[i:], op.Arg2.Val)
				i += 2
			}
		} else if op.Arg1.Code.Immediate() {
			bo.PutUint16(buf[i:], op.Arg1.Val)
			i += 2
		}
	}
	return i
}

// DecodeOp decodes an Op from buf at the given index, returning it and the
// index of the next un-decoded byte.  Decode panics if the buffer isn't large
// enough (ends in a truncated operation).
func DecodeOp(bo binary.ByteOrder, buf []byte, i int) (Op, int) {
	var op Op
	op.Code = OpCode(buf[i])
	i++
	switch op.Code.Arity() {
	case 1:
		op.Arg1.Code = ArgCode(buf[i])
		i++
		if op.Arg1.Code.Immediate() {
			op.Arg1.Val = bo.Uint16(buf[i:])
			i += 2
		}
	case 2:
		op.Arg1.Code = ArgCode(buf[i])
		i++
		op.Arg2.Code = ArgCode(buf[i])
		i++
		if op.Arg1.Code.Immediate() {
			op.Arg1.Val = bo.Uint16(buf[i:])
			i += 2
		}
		if op.Arg2.Code.Immediate() {
			op.Arg2.Val = bo.Uint16(buf[i:])
			i += 2
		}
	}
	return op, i
}
