package stackvm

import "errors"

const _size = 256

var (
	errStackOverflow  = errors.New("stack overflow")
	errStackUnderflow = errors.New("stack underflow")
	errOutOfMemory    = errors.New("out of memory")
	errSegfault       = errors.New("segfault")
	errPCRange        = errors.New("PC out of range")
	errCannotFork     = errors.New("cannot fork")
	errFullQ          = errors.New("refusing to fork, queue full")
)

// TODO: fork

// Mach is a stack machine
type Mach struct {
	prog  []Op
	emit  func(*Mach) error
	err   error
	pc    int
	heap  int
	stack int
	mem   [_size]byte
}

// Op is a stack machine operation.
type Op interface {
	run(*Mach) error
}

// Prog returns a copy of the compiled program.
func (m *Mach) Prog() []Op {
	return append([]Op(nil), m.prog...)
}

// Err returns any runtime error.
func (m *Mach) Err() error {
	return m.err
}

// PC returns the index, in Prog(), of the next Op to be executed.
func (m *Mach) PC() int {
	return m.pc
}

// Stack returns a copy of the current stack.
func (m *Mach) Stack() []byte {
	return append([]byte(nil), m.mem[0:m.stack]...)
}

// Heap returns a copy of the current heap.
func (m *Mach) Heap() []byte {
	return append([]byte(nil), m.mem[m.heap:]...)
}

// Run runs the current machine until an error occurs or the program
// terminates.
func (m *Mach) Run() error {
	if m.err != nil {
		return m.err
	}
	for m.pc < len(m.prog) {
		op := m.prog[m.pc]
		m.pc++
		m.err = op.run(m)
		if m.err != nil {
			break
		}
	}
	return m.err
}

type runq struct {
	q []*Mach
}

func (rq *runq) emit(n *Mach) error {
	if len(rq.q) == cap(rq.q) {
		return errFullQ
	}
	rq.q = append(rq.q, n)
	return nil
}

func (rq *runq) next() *Mach {
	if len(rq.q) == 0 {
		return nil
	}
	m := rq.q[0]
	rq.q = rq.q[:copy(rq.q, rq.q[1:])]
	return m
}

// RunAll runs the current machine and any forked copies until error or program
// termination. Pending machines are kept on a fixed-size queue; when this
// queue is full, forks will fail. After a machine terminates, the given emit
// function is called on it. If the emit function returns non-nil error, that
// error is returned immediately from RunAll and no further machines are ran.
func (m *Mach) RunAll(maxPending int, emit func(*Mach) error) error {
	if m.err != nil {
		return m.err
	}
	rq := runq{make([]*Mach, 0, maxPending)}
	m.emit = rq.emit
	for m != nil {
		m.Run()
		if err := emit(m); err != nil {
			return err
		}
		m = rq.next()
	}
	return nil
}

// Tracer observes stack machine execution.
type Tracer interface {
	Begin(m *Mach)
	End(m *Mach, err error)
	Before(m *Mach, pc int, op Op)
	Fork(m, n *Mach, pc int, next Op)
	After(m *Mach, pc int, last, next Op, err error)
}

// Trace runs the current machine like Run while calling the given Tracer.
func (m *Mach) Trace(t Tracer) (err error) {
	if m.err != nil {
		return m.err
	}
	t.Begin(m)
	for m.pc < len(m.prog) {
		op := m.prog[m.pc]
		t.Before(m, m.pc, op)
		m.pc++
		m.err = op.run(m)
		if m.pc < len(m.prog) {
			t.After(m, m.pc, op, m.prog[m.pc], m.err)
		} else {
			t.After(m, m.pc, op, nil, m.err)
		}
		if m.err != nil {
			break
		}
	}
	t.End(m, m.err)
	return m.err
}

func (m *Mach) jump(off int) error {
	i := m.pc + off
	if i < 0 || i > len(m.prog) {
		return errPCRange
	}
	m.pc = i
	return nil
}

func (m *Mach) copy() *Mach {
	n := *m
	return &n
}

func (m *Mach) fork(off int) error {
	i := m.pc + off
	if i < 0 || i > len(m.prog) {
		return errPCRange
	}
	if m.emit == nil {
		return errCannotFork
	}
	n := m.copy()
	n.pc = i
	return m.emit(n)
}

func (m *Mach) ref(off int) (i int, err error) {
	i = m.stack - off - 1
	if i < 0 {
		err = errStackUnderflow
	}
	return
}

func (m *Mach) pop() (val byte, err error) {
	if i := m.stack - 1; i < 0 {
		err = errStackUnderflow
	} else {
		val = m.mem[i]
		m.stack = i
	}
	return
}

func (m *Mach) push(val byte) (err error) {
	if i := m.stack; i >= m.heap {
		err = errStackOverflow
	} else {
		m.mem[i] = val
		m.stack++
	}
	return
}

func (m *Mach) alloc(n int) (err error) {
	i := m.heap - n
	if i < m.stack {
		return errOutOfMemory
	}
	m.heap = i
	return nil
}

func (m *Mach) load(off int) (val byte, err error) {
	i := _size - off
	if i < m.heap || i > _size {
		return 0, errSegfault
	}
	return m.mem[i], nil
}

func (m *Mach) store(off int, val byte) (err error) {
	i := _size - off
	if i < m.heap || i > _size {
		return errSegfault
	}
	m.mem[i] = val
	return nil
}
