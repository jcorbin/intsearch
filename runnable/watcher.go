package runnable

// SearchWatcher is the interface implemented to observe a search run.
type SearchWatcher interface {
	BeforeStep(srch Searcher, sol *Solution)
	Stepped(srch Searcher, sol *Solution)
	Emitted(srch Searcher, child *Solution)
}

// Watchers allows combining more than one SearchWatcher.
type Watchers []SearchWatcher

// BeforeStep calls each watcher BeforeStep.
func (ws Watchers) BeforeStep(srch Searcher, sol *Solution) {
	for _, w := range ws {
		w.BeforeStep(srch, sol)
	}
}

// Stepped calls each watcher Stepped.
func (ws Watchers) Stepped(srch Searcher, sol *Solution) {
	for _, w := range ws {
		w.Stepped(srch, sol)
	}
}

// Emitted calls each watcher Emitted.
func (ws Watchers) Emitted(srch Searcher, child *Solution) {
	for _, w := range ws {
		w.Emitted(srch, child)
	}
}
