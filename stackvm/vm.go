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

// Tracer observes stack machine execution.
type Tracer interface {
	Begin(m *Mach)
	End(m *Mach, err error)
	Before(m *Mach, pc int, op Op)
	Fork(m, n *Mach, pc int, next Op)
	After(m *Mach, pc int, last, next Op, err error)
}

// Mach is a stack machine
type Mach struct {
	prog  []Op
	ctx   context
	err   error
	pc    int
	heap  int
	stack int
	mem   [_size]byte
}

// Handler is a result machine handler, for use with the Fork family of
// operations. Handler values can be registered with Mach.Handle.
type Handler interface {
	Handle(*Mach) error
}

type context interface {
	Handler
	queue(*Mach) error
	next() *Mach
}

// Op is a stack machine operation.
type Op interface {
	run(*Mach) error
}

// Prog returns a copy of the compiled program.
func (m *Mach) Prog() []Op {
	return append([]Op(nil), m.prog...)
}

// Handle sets a result handling function, without which the Fork family of
// operations cannot work. When a handle func has been set, a queue of pending
// machines is built up by fork operations. This queue has fixed size, the
// maxPending argument here, after which fork operations will fail.
//
// The hanlde func gets called on every machine, original or copy, after it has
// terminated (with or without an error). Any machine error, including perhaps
// failure to fork due to queue being full, is available using the Mach.Err()
// method. The handle func may decide what to do about any such error: if it
// returns non-nil, then the machine run stops and returns that error
// immediately; otherwise the run continues, calling handle with any further
// results.
func (m *Mach) Handle(maxPending int, h Handler) {
	q := runq{make([]*Mach, 0, maxPending)}
	m.ctx = handler{h, &q}
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
	if m.ctx == nil {
		m._run()
		return m.err
	}
	for m != nil {
		m._run()
		if err := m.ctx.Handle(m); err != nil {
			return err
		}
		m = m.ctx.next()
	}
	return nil
}

func (m *Mach) _run() {
	for m.err == nil && m.pc < len(m.prog) {
		op := m.prog[m.pc]
		m.pc++
		m.err = op.run(m)
	}
}

// Trace runs the current machine like Run while calling the given Tracer.
func (m *Mach) Trace(t Tracer) (err error) {
	if m.err != nil {
		return m.err
	}
	if m.ctx == nil {
		m._trace(t)
		return m.err
	}
	traceContext(m, t)
	for m != nil {
		m._trace(t)
		if err := m.ctx.Handle(m); err != nil {
			return err
		}
		m = m.ctx.next()
	}
	return nil
}

func (m *Mach) _trace(t Tracer) {
	t.Begin(m)
	for m.err == nil && m.pc < len(m.prog) {
		op := m.prog[m.pc]
		t.Before(m, m.pc, op)
		m.pc++
		m.err = op.run(m)
		if m.pc < len(m.prog) {
			t.After(m, m.pc, op, m.prog[m.pc], m.err)
		} else {
			t.After(m, m.pc, op, nil, m.err)
		}
	}
	t.End(m, m.err)
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
	if m.ctx == nil {
		return errCannotFork
	}
	n := m.copy()
	n.pc = i
	return m.ctx.queue(n)
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

type runq struct{ q []*Mach }

func (rq *runq) queue(n *Mach) error {
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

type handler struct {
	Handler
	*runq
}

func traceContext(m *Mach, t Tracer) {
	ctx := m.ctx
	if ft, ok := ctx.(forkTracer); ok {
		ctx = ft.context
	}
	m.ctx = forkTracer{ctx, m, t}
}

type forkTracer struct {
	context
	m *Mach
	t Tracer
}

func (ft forkTracer) queue(n *Mach) error {
	traceContext(n, ft.t)
	if err := ft.context.queue(n); err != nil {
		return err
	}
	ft.t.Fork(ft.m, n, n.pc, n.prog[n.pc])
	return nil
}

// HandleFunc is a convenience for implementing Handler.
type HandleFunc func(*Mach) error

// JustHandleFunc is a convenience for implementing a simple result handler.
type JustHandleFunc func(*Mach)

// Handle calls the underlying func.
func (f HandleFunc) Handle(m *Mach) error { return f(m) }

// Handle first checks for, and returns any non-nil Mach.Err(). Passing that,
// it calls the underlying func(m), and returns nil.
func (f JustHandleFunc) Handle(m *Mach) error {
	if err := m.Err(); err != nil {
		return err
	}
	f(m)
	return nil
}
