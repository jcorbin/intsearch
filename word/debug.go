package word

import (
	"fmt"

	"github.com/jcorbin/intsearch/internal"
)

type debugID [3]uint64

func rootDebugID(id uint64) debugID {
	if id == 0 {
		panic("invalid root debug id")
	}
	return debugID{id, 0, id}
}

func (di debugID) SubID(id uint64) debugID {
	if id == 0 {
		panic("invalid sub debug id")
	}
	return debugID{di[0], di[2], id}
}

func (di debugID) String() string {
	return fmt.Sprintf("%v:%v:%v", di[0], di[1], di[2])
}

// debugger implements a Watcher that manages an id mapping ids for each
// outstanding solution.
type debugger struct {
	id   uint64
	idOf map[Solution]debugID
}

func (dbg *debugger) getOrAddID(parent, child Solution) debugID {
	id, defined := dbg.idOf[child]
	if !defined {
		if parent == nil {
			dbg.id++
			id = rootDebugID(dbg.id)
		} else {
			parID := dbg.getOrAddID(nil, parent)
			dbg.id++
			id = parID.SubID(dbg.id)
		}
		dbg.idOf[child] = id
	}
	return id
}

// Result removes any id mapping.
func (dbg *debugger) Result(sol Solution) bool {
	delete(dbg.idOf, sol)
	return false
}

// Before adds an id mapping, if none exists.
func (dbg *debugger) Before(sol Solution) {
	dbg.getOrAddID(nil, sol)
}

// After adds an id mapping, if none exists.
func (dbg *debugger) After(sol Solution) {
	dbg.getOrAddID(nil, sol)
}

// Fork adds an id mapping for the new child.
func (dbg *debugger) Fork(parent, child Solution) {
	dbg.getOrAddID(parent, child)
}

// DebugWatcher implements a Watcher that prints a debug log.
type DebugWatcher struct {
	debugger
	Logf func(string, ...interface{})
}

// NewDebugWatcher returns a new debug watcher.
func NewDebugWatcher(logf func(string, ...interface{})) *DebugWatcher {
	return &DebugWatcher{
		debugger: debugger{
			idOf: make(map[Solution]debugID),
		},
		Logf: logf,
	}
}

// Result prints a "=== ..." line.
func (dbg *DebugWatcher) Result(sol Solution) bool {
	sol.Dump(internal.ElidedF(dbg.Logf,
		fmt.Sprintf("=== %v", dbg.getOrAddID(nil, sol))))
	delete(dbg.idOf, sol)
	return false
}

// Before prints a "--> ..." line.
func (dbg *DebugWatcher) Before(sol Solution) {
	sol.Dump(internal.ElidedF(dbg.Logf,
		fmt.Sprintf("--> %v", dbg.getOrAddID(nil, sol))))
}

// After prints a "<-- ..." line.
func (dbg *DebugWatcher) After(sol Solution) {
	sol.Dump(internal.ElidedF(dbg.Logf,
		fmt.Sprintf("<-- %v", dbg.getOrAddID(nil, sol))))
}

// Fork prints a "+++ ..." line.
func (dbg *DebugWatcher) Fork(parent, child Solution) {
	child.Dump(internal.ElidedF(dbg.Logf,
		fmt.Sprintf("+++ %v", dbg.getOrAddID(parent, child))))
}
