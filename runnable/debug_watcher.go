package runnable

// DebugWatcher simply prints a debug log of the search run.
type DebugWatcher struct {
	Logf func(string, ...interface{})
}

// Emitted simply prints the newlly emitted solution, noting the frontier size
// and any parent.
func (wat DebugWatcher) Emitted(srch Searcher, child *Solution) {
	wat.Logf("+++ %v %v", srch.FrontierSize(), child)
	if parent := srch.Current(); parent != nil {
		wat.Logf("... parent %v", parent)
	}
}

// BeforeStep prints each solution before it is stepped.
func (wat DebugWatcher) BeforeStep(srch Searcher, sol *Solution) {
	wat.Logf(">>> %v", sol)
}

// Stepped prints each solution after it has been stepped.  Furthemore:
// - if the last step stored a new letter, the new mapping is printed.
// - if the last step was a fork, then the frontier size is printed.
func (wat DebugWatcher) Stepped(srch Searcher, sol *Solution) {
	wat.Logf("... %v", sol)
	if isStoreStep(sol.steps[sol.stepi-1]) {
		wat.Logf("... %s", sol.LetterMapping())
	} else if isForkStep(sol.steps[sol.stepi-1]) {
		wat.Logf("... len(frontier) == %v", srch.FrontierSize())
	}
}
