package main

import "errors"

var (
	errStackOverflow  = errors.New("stack overflow")
	errStackUnderflow = errors.New("stack underflow")
	errSegfault       = errors.New("stack segfault")
	errOutOfMemory    = errors.New("out of memory")
	errFrontierFull   = errors.New("frontier full")
)

const _machSize = 256

type machStep interface {
	step(m *mach) error
}

type mach struct {
	emit            func(*mach) error
	prog            []machStep
	mem             [_machSize]int
	pc, heap, stack int
}

type machSearch struct {
	frontier chan *mach
}

func (s machSearch) emit(m *mach) error {
	select {
	case s.frontier <- m:
		return nil
	default:
		return errFrontierFull
	}
}

func runSearch(prog []machStep, emit func(m *mach) bool) error {
	s := machSearch{
		frontier: make(chan *mach, 1024),
	}
	s.frontier <- &mach{
		emit: s.emit,
		prog: prog,
		heap: _machSize,
	}
	var err error
	for {
		select {
		case m := <-s.frontier:
			err = m.run()
			if err != nil {
				continue
			}
			if emit(m) {
				return nil
			}
		default:
			return err
		}
	}
}

func (m *mach) run() error {
	for m.pc < len(m.prog) {
		ms := m.prog[m.pc]
		m.pc++
		if err := ms.step(m); err != nil {
			return err
		}
	}
	return nil
}

func (m *mach) fork(offset int) error {
	newm := *m
	newm.pc += offset
	return m.emit(&newm)
}

func (m *mach) needStack(n int) error {
	if m.stack <= n-1 {
		return errStackUnderflow
	}
	return nil
}

func (m *mach) addr(offset int) (int, error) {
	addr := m.heap - offset
	if addr < m.stack || addr >= len(m.mem) {
		return 0, errSegfault
	}
	return addr, nil
}

func (m *mach) peek() (int, error) {
	if m.stack <= 0 {
		return 0, errStackUnderflow
	}
	return m.mem[m.stack], nil
}

func (m *mach) pop() (int, error) {
	if m.stack <= 0 {
		return 0, errStackUnderflow
	}
	m.stack--
	return m.mem[m.stack], nil
}

func (m *mach) push(val int) error {
	i := m.stack
	if i >= m.heap {
		return errStackOverflow
	}
	m.stack++
	m.mem[i] = val
	return nil
}

func (op push) step(m *mach) error { return m.push(int(op)) }
func (op _dup) step(m *mach) error {
	val, err := m.peek()
	if err != nil {
		return err
	}
	return m.push(val)
}
func (op _swap) step(m *mach) error {
	if err := m.needStack(2); err != nil {
		return err
	}
	i := m.stack - 1
	j := i - 1
	m.mem[i], m.mem[j] = m.mem[j], m.mem[i]
	return nil
}
func (op _alloc) step(m *mach) error {
	offset, err := m.pop()
	if err != nil {
		return err
	}
	nh := m.heap - offset
	if nh <= m.stack {
		return errOutOfMemory
	}
	return nil
}
func (op _load) step(m *mach) error {
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	offset := m.mem[i]
	addr, err := m.addr(offset)
	if err != nil {
		return err
	}
	m.mem[i] = m.mem[addr]
	return nil
}
func (op _store) step(m *mach) error {
	offset, err := m.pop()
	if err != nil {
		return err
	}
	val, err := m.pop()
	if err != nil {
		return err
	}
	addr, err := m.addr(offset)
	if err != nil {
		return err
	}
	m.mem[addr] = val
	return nil
}

func (op halt) step(m *mach) error { return op.error }
func (op jmp) step(m *mach) error {
	m.pc += int(op)
	return nil
}
func (op jnz) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		m.pc += int(op)
	}
	return nil
}
func (op jz) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		m.pc += int(op)
	}
	return nil
}
func (op fork) step(m *mach) error {
	return m.fork(int(op))
}
func (op fnz) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.fork(int(op))
	}
	return nil
}
func (op fz) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		return m.fork(int(op))
	}
	return nil
}
func (op _add) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	m.mem[i] += val
	return nil
}
func (op _sub) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	m.mem[i] -= val
	return nil
}
func (op _div) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	m.mem[i] /= val
	return nil
}
func (op _mod) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	m.mem[i] %= val
	return nil
}
func (op _lt) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	if m.mem[i] < val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}
func (op _lte) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	if m.mem[i] <= val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}
func (op _eq) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	if m.mem[i] == val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}
func (op _neq) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	if m.mem[i] != val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}
func (op _gt) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	if m.mem[i] > val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}
func (op _gte) step(m *mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if err := m.needStack(1); err != nil {
		return err
	}
	i := m.stack - 1
	if m.mem[i] <= val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}
