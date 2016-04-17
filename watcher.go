package main

type searchWatcher interface {
	beforeStep(srch searcher, sol *solution)
	stepped(srch searcher, sol *solution)
	emitted(srch searcher, child *solution)
}

type watchers []searchWatcher

func (ws watchers) beforeStep(srch searcher, sol *solution) {
	for _, w := range ws {
		w.beforeStep(srch, sol)
	}
}

func (ws watchers) stepped(srch searcher, sol *solution) {
	for _, w := range ws {
		w.stepped(srch, sol)
	}
}

func (ws watchers) emitted(srch searcher, child *solution) {
	for _, w := range ws {
		w.emitted(srch, child)
	}
}
