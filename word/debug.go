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

// DebugWatcher implements a Watcher that prints a debug log.
type DebugWatcher struct {
	Logf func(string, ...interface{})
	id   uint64
	idOf map[Solution]debugID
}

// NewDebugWatcher returns a new debug watcher.
func NewDebugWatcher(logf func(string, ...interface{})) *DebugWatcher {
	return &DebugWatcher{
		Logf: logf,
		idOf: make(map[Solution]debugID),
	}
}

func (dbg *DebugWatcher) getOrAddID(parent, child Solution) debugID {
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

// Result prints a "=== ..." line.
func (dbg *DebugWatcher) Result(sol Solution) bool {
	id := dbg.getOrAddID(nil, sol)
	sol.Dump(internal.PrefixedF(dbg.Logf,
		fmt.Sprintf("=== %v", id),
		fmt.Sprintf("... %v", id)))
	delete(dbg.idOf, sol)
	return false
}

// Before prints a "--> ..." line.
func (dbg *DebugWatcher) Before(sol Solution) {
	id := dbg.getOrAddID(nil, sol)
	sol.Dump(internal.PrefixedF(dbg.Logf,
		fmt.Sprintf("--> %v", id),
		fmt.Sprintf("... %v", id)))
}

// After prints a "<-- ..." line.
func (dbg *DebugWatcher) After(sol Solution) {
	id := dbg.getOrAddID(nil, sol)
	sol.Dump(internal.PrefixedF(dbg.Logf,
		fmt.Sprintf("<-- %v", id),
		fmt.Sprintf("... %v", id)))
}

// Fork prints a "+++ ..." line.
func (dbg *DebugWatcher) Fork(parent, child Solution) {
	id := dbg.getOrAddID(parent, child)
	child.Dump(internal.PrefixedF(dbg.Logf,
		fmt.Sprintf("+++ %v", id),
		fmt.Sprintf("... %v", id)))
}
