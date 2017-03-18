package main

import (
	"errors"
	"fmt"
)

type machStep interface {
	step(*mach) error
}

const _machSize = 256

type mach struct {
	prog        []machStep
	pc          int
	stack, heap int
	mem         [_machSize]byte
}

func newMach(prog []machStep) *mach {
	return &mach{
		prog: prog,
		heap: _machSize,
	}
}

func (m *mach) run() error {
	for m.pc < len(m.prog) {
		ms := m.prog[m.pc]
		fmt.Printf("STEP %03d: % 10v", m.pc, ms)
		m.pc++
		if err := ms.step(m); err != nil {
			fmt.Printf(" -- ERR: %v\n", err)
			return err
		}
		fmt.Printf(" -- %v %v\n",
			m.mem[m.heap:],
			m.mem[:m.stack],
		)
	}
	return nil
}

var (
	errUnimplemented  = errors.New("op unimplemented")
	errStackOverflow  = errors.New("stack overflow")
	errStackUnderflow = errors.New("stack underflow")
	errOutOfMemory    = errors.New("out of memory")
	errSegfault       = errors.New("segfault")
)

func (m *mach) push(v byte) error {
	i := m.stack
	if i >= m.heap {
		return errStackOverflow
	}
	m.stack = i + 1
	m.mem[i] = v
	return nil
}

func (m *mach) peek() (byte, error) {
	i := m.stack - 1
	if i < 0 {
		return 0, errStackUnderflow
	}
	return m.mem[i], nil
}

func (m *mach) pop() (byte, error) {
	i := m.stack - 1
	if i < 0 {
		return 0, errStackUnderflow
	}
	v := m.mem[i]
	m.stack = i
	return v, nil
}

func (op push) step(m *mach) error {
	return m.push(byte(op))
}

func (op _add) step(m *mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a + b)
}

func (op _sub) step(m *mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a - b)
}

func (op _mod) step(m *mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a % b)
}

func (op _div) step(m *mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	return m.push(a / b)
}

func (op _alloc) step(m *mach) error {
	v, err := m.pop()
	if err != nil {
		return err
	}
	i := m.heap - int(v)
	if i < m.stack {
		return errOutOfMemory
	}
	m.heap = i
	return nil
}

func (op _lt) step(m *mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a < b {
		return m.push(1)
	}
	return m.push(0)
}

func (op _eq) step(m *mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if a == b {
		return m.push(1)
	}
	return m.push(0)
}

func (op fnz) step(m *mach) error {
	v, err := m.pop()
	if err != nil {
		return err
	}
	if v != 0 {
		// TODO: fork int(op)
	}
	return nil
}

func (op jmp) step(m *mach) error {
	m.pc += int(op)
	return nil
}

func (op jnz) step(m *mach) error {
	v, err := m.pop()
	if err != nil {
		return err
	}
	if v != 0 {
		m.pc += int(op)
	}
	return nil
}

func (op jz) step(m *mach) error {
	v, err := m.pop()
	if err != nil {
		return err
	}
	if v == 0 {
		m.pc += int(op)
	}
	return nil
}

func (op _dup) step(m *mach) error {
	v, err := m.peek()
	if err != nil {
		return err
	}
	return m.push(v)
}

func (op _swap) step(m *mach) error {
	b, err := m.pop()
	if err != nil {
		return err
	}
	a, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.push(a); err != nil {
		return err
	}
	return m.push(b)
}

func (m *mach) load(off int) (byte, error) {
	if off < 0 {
		return 0, errSegfault
	}
	i := m.heap + off
	if i < m.stack || i >= _machSize {
		return 0, errSegfault
	}
	return m.mem[i], nil
}

func (op _load) step(m *mach) error {
	v, err := m.peek()
	if err != nil {
		return err
	}
	mv, err := m.load(int(v))
	if err != nil {
		return err
	}
	m.mem[m.stack-1] = mv
	return nil
}
func (op _store) step(_ *mach) error { return errUnimplemented }

func (op _halt) step(_ *mach) error { return op.err }

func (op _hnz) step(m *mach) error {
	v, err := m.pop()
	if err != nil {
		return err
	}
	if v != 0 {
		return op.err
	}
	return nil
}

func (op _hz) step(m *mach) error {
	v, err := m.pop()
	if err != nil {
		return err
	}
	if v == 0 {
		return op.err
	}
	return nil
}

// func (op label) step(_ *mach) error    { return errUnimplemented }
// func (op labelRef) step(_ *mach) error { return errUnimplemented }
