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
