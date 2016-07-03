package word

// Resultor consumes results from a Plan.
type Resultor interface {
	// Result must extract any desired information from the the passed
	// solution, and must not retain a reference to it.  Result should return
	// true only if it wishes the search to stop (e.g. stop on first solution).
	Result(sol Solution) bool
}

// ResultFunc is a convenience resultor.
type ResultFunc func(Solution) bool

// Result calls the wrapped function.
func (rf ResultFunc) Result(sol Solution) bool {
	return rf(sol)
}

// Watcher is an optional interface that a Resultor may implement to
// observe the plan execution trace.
type Watcher interface {
	Resultor
	Before(sol Solution)
	After(sol Solution)
	Fork(parent, child Solution)
}

// ResultWrapper is an optional interface that a Watcher may implement to wrap
// results under a MultiWatcher.  For example, it is used by TraceWatcher to
// associate traces with solutions.
type ResultWrapper interface {
	Watcher
	WrapResult(sol Solution) Solution
}

// ResultWatcher converts a Resultor into a Watcher with noop Before/After/Fork
// methods for use in a MultiWatcher.
type ResultWatcher struct {
	Resultor
}

// Before does nothing.
func (rw ResultWatcher) Before(_ Solution) {
}

// After does nothing.
func (rw ResultWatcher) After(_ Solution) {
}

// Fork does nothing.
func (rw ResultWatcher) Fork(_, _ Solution) {
}

// MultiWatcher simply passes each Watcher event to each watcher.
type MultiWatcher []Watcher

// Watchers is a convenience constructor for a MultiWatcher.
func Watchers(ws ...Watcher) MultiWatcher {
	return MultiWatcher(ws)
}

// Result passes the result to each watcher, and returns true if any of them
// do.
func (ws MultiWatcher) Result(sol Solution) (r bool) {
	for _, w := range ws {
		if rw, ok := w.(ResultWrapper); ok {
			sol = rw.WrapResult(sol)
		} else {
			r = w.Result(sol) || r
		}
	}
	return
}

// Before calls each watcher before method.
func (ws MultiWatcher) Before(sol Solution) {
	for _, w := range ws {
		w.Before(sol)
	}
}

// After calls each watcher after method.
func (ws MultiWatcher) After(sol Solution) {
	for _, w := range ws {
		w.After(sol)
	}
}

// Fork calls each watcher fork method.
func (ws MultiWatcher) Fork(parent, child Solution) {
	for _, w := range ws {
		w.Fork(parent, child)
	}
}

// MetricWatcher implements a Watcher that counts.
type MetricWatcher struct {
	// TODO: better metric support
	Results int
	Steps   int
	Forks   int
}

// Result increments the Results counter.
func (met *MetricWatcher) Result(sol Solution) bool {
	met.Results++
	return false
}

// Before does nothing.
func (met *MetricWatcher) Before(sol Solution) {
}

// After increments the Steps counter.
func (met *MetricWatcher) After(sol Solution) {
	met.Steps++
}

// Fork increments the Forks counter and MaxFrontierLen gauge.
func (met *MetricWatcher) Fork(parent, child Solution) {
	met.Forks++
}
