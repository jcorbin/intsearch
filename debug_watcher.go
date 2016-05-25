package main

type debugWatcher struct {
	logf func(string, ...interface{})
}

func (wat debugWatcher) emitted(srch searcher, child *solution) {
	wat.logf("+++ %v %v", srch.frontierSize(), child)
	if parent := srch.current(); parent != nil {
		wat.logf("... parent %v", parent)
	}
}

func (wat debugWatcher) beforeStep(srch searcher, sol *solution) {
	wat.logf(">>> %v", sol)
}

func (wat debugWatcher) stepped(srch searcher, sol *solution) {
	wat.logf("... %v", sol)
	if isStoreStep(sol.steps[sol.stepi-1]) {
		wat.logf("... %s", sol.letterMapping())
	} else if isForkStep(sol.steps[sol.stepi-1]) {
		wat.logf("... len(frontier) == %v", srch.frontierSize())
	}
}
