package opcode

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Machine is the interface implemented by any opcode machine.
type Machine interface {
	ByteOrder() binary.ByteOrder
	Reset()
	Load(p []byte, assumeValid bool) error
	LastOp() Op
	NextOp() Op
	Step()
	Run()
	RunAll(res Resultor) Machine
	Program() []byte
	PI() int
	State() interface{}
	CopyMemory(offset int, buf []byte) int
	Check() error
}

var (
	// ErrInvalidPI is returned by Machine.Check if the program index (pi) is
	// invalid (beyond the length of the program).
	ErrInvalidPI = errors.New("invalid PI")

	// ErrNotHalted is returned by Machine.Check if the machine has not yet
	// halted.
	ErrNotHalted = errors.New("not halted")
)

// ErrOpLimitExceeded means that the machine exceeded its specified operation
// count limit.
type ErrOpLimitExceeded struct {
	cnt, lim uint
}

func (ole ErrOpLimitExceeded) Error() string {
	return fmt.Sprintf("op count limit exceeded: %d >= %d", ole.cnt, ole.lim)
}

// ErrNonZeroHalt means that the machine halted with a non-zero code.
type ErrNonZeroHalt byte

func (nz ErrNonZeroHalt) Error() string {
	return fmt.Sprintf("halted with non-zero return code 0x%02x", nz)
}

// Resultor is the interface required a machine result consumer.
type Resultor interface {
	Result(mach Machine) bool
}

// ResultFunc is a convenience resultor.
type ResultFunc func(Machine) bool

// Result calls the wrapped function.
func (rf ResultFunc) Result(mach Machine) bool {
	return rf(mach)
}

// Tracer is an optional interface that a Resultor may implement to observe the
// full execution trace.
type Tracer interface {
	Resultor
	Before(mach Machine)
	After(mach Machine)
	Emit(action string, parent, child Machine)
}

// Tracers is a convenience constructor for a MultiTracer.
func Tracers(ts ...Tracer) MultiTracer {
	return MultiTracer(ts)
}

// MultiTracer dispatches to more than one Tracer.
type MultiTracer []Tracer

// Result calls all Results.
func (ts MultiTracer) Result(mach Machine) (r bool) {
	for _, t := range ts {
		r = t.Result(mach) || r
	}
	return
}

// Before calls all Befores.
func (ts MultiTracer) Before(mach Machine) {
	for _, t := range ts {
		t.Before(mach)
	}
}

// After calls all Afters.
func (ts MultiTracer) After(mach Machine) {
	for _, t := range ts {
		t.After(mach)
	}
}

// Emit calls all Emits.
func (ts MultiTracer) Emit(action string, parent, child Machine) {
	for _, t := range ts {
		t.Emit(action, parent, child)
	}
}

// TracePrinter is a Tracer that prints machine data at every turn.
type TracePrinter struct {
	ids map[Machine]int
}

// NewTracePrinter creates a new trace printer.
func NewTracePrinter() *TracePrinter {
	return &TracePrinter{ids: make(map[Machine]int)}
}

// Result prints the result machine state, including a dump of memory.
func (tp *TracePrinter) Result(mach Machine) bool {
	id := tp.ids[mach]
	if err := mach.Check(); err != nil {
		fmt.Printf("[% 3d]  ! ERROR: %v\n", id, err)
	}
	fmt.Printf("[% 3d]  = done: %v\n", id, mach)
	dumpMemory(mach, func(format string, args ...interface{}) {
		format = fmt.Sprintf("		 %s\n", format)
		fmt.Printf(format, args...)
	})
	return false
}

// Before prints the machine state and operation about to be executed.
func (tp *TracePrinter) Before(mach Machine) {
	id := tp.ids[mach]
	fmt.Printf("[% 3d]--> % -50v // NEXT: %v\n", id, mach.State(), mach.NextOp())
}

// After prints the operation that was just ran, and the resulting machine
// state.
func (tp *TracePrinter) After(mach Machine) {
	id := tp.ids[mach]
	fmt.Printf("[% 3d]  = % -50v // LAST: %v\n", id, mach.State(), mach.LastOp())
}

// Emit prints the machine state and operation that has been deferred.
func (tp *TracePrinter) Emit(action string, parent, child Machine) {
	id := len(tp.ids) + 1
	tp.ids[child] = id
	fmt.Printf("[% 3d]  X % -50d // NEXT: %v\n", id, child.State(), child.NextOp())
}

func dumpMemory(mach Machine, outf func(string, ...interface{})) {
	var (
		m [16]byte
		o int
	)
	for {
		switch mach.CopyMemory(o, m[:]) {
		case 0:
			return
		case 1:
			outf("%04x: %02x", o, m[0])
			o++
		case 2:
			outf("%04x: %02x%02x ", o, m[0], m[1])
			o += 2
		case 3:
			outf("%04x: %02x%02x %02x", o, m[0], m[1], m[2])
			o += 3
		case 4:
			outf("%04x: %02x%02x %02x%02x ", o, m[0], m[1], m[2], m[3])
			o += 4
		case 5:
			outf("%04x: %02x%02x %02x%02x %02x", o, m[0], m[1], m[2], m[3], m[4])
			o += 5
		case 6:
			outf("%04x: %02x%02x %02x%02x %02x%02x ", o, m[0], m[1], m[2], m[3], m[4], m[5])
			o += 6
		case 7:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6])
			o += 7
		case 8:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x ", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7])
			o += 8
		case 9:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8])
			o += 9
		case 10:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x ", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9])
			o += 10
		case 11:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10])
			o += 11
		case 12:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x ", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11])
			o += 12
		case 13:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11], m[12])
			o += 13
		case 14:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x ", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11], m[12], m[13])
			o += 14
		case 15:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11], m[12], m[13], m[14])
			o += 15
		case 16:
			outf("%04x: %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x %02x%02x", o, m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7], m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15])
			o += 16
		default:
			panic("invalid CopyMemory return count")
		}
	}
}
