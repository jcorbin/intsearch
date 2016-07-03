package word

import (
	"fmt"

	"github.com/jcorbin/intsearch/internal"
)

// TracedSolution is a captured solution with a trace.
type TracedSolution struct {
	CapturedSolution
	trace []string
}

// Trace returns the captured trace.
func (ts *TracedSolution) Trace() []string {
	return ts.trace
}

// Dump emits the captured dump plus the collected trace.
func (ts *TracedSolution) Dump(logf func(string, ...interface{})) {
	ts.CapturedSolution.Dump(logf)
	for _, s := range ts.trace {
		logf(s)
	}
}

// TraceWatcher implements a Watcher that collects solution traces.
type TraceWatcher struct {
	debugger
	Trace map[Solution][]string
}

// NewTraceWatcher creates a new trace watcher.
func NewTraceWatcher() *TraceWatcher {
	return &TraceWatcher{
		debugger: debugger{
			idOf: make(map[Solution]debugID),
		},
		Trace: make(map[Solution][]string),
	}
}

// Result removes the associated trace.
func (trc *TraceWatcher) Result(sol Solution) bool {
	delete(trc.idOf, sol)
	delete(trc.Trace, sol)
	return false
}

// WrapResult captures the solution and attaches the collected trace.
func (trc *TraceWatcher) WrapResult(sol Solution) Solution {
	t := &TracedSolution{
		trace: trc.takeDump(sol),
	}
	t.Capture(sol)
	delete(trc.Trace, sol)
	return t
}

// Before adds a dump to the solution trace.
func (trc *TraceWatcher) Before(sol Solution) {
	trc.takeDump(sol)
}

func (trc *TraceWatcher) takeDump(sol Solution) []string {
	id := trc.getOrAddID(nil, sol)
	t := trc.Trace[sol]
	sol.Dump(internal.ElidedF(
		func(format string, args ...interface{}) {
			t = append(t, fmt.Sprintf(format, args...))
		},
		fmt.Sprintf("%04d %v>", len(t), id)))
	trc.Trace[sol] = t
	return t
}

// Fork copies the parent trace for the child, and adds a fork marker to both.
func (trc *TraceWatcher) Fork(parent, child Solution) {
	pid := trc.getOrAddID(nil, parent)
	cid := trc.getOrAddID(parent, child)

	t := trc.Trace[parent]
	i := len(t)
	t = append(t, fmt.Sprintf("%04d %v> forked %v", i, pid, cid))
	trc.Trace[parent] = t

	t = append([]string(nil), t...)
	t[len(t)-1] = fmt.Sprintf("%04d %v> forked from %v", i, cid, pid)
	trc.Trace[child] = t
}
