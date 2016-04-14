package main

type searchWatcher interface {
	beforeStep(srch searcher, sol *solution)
	stepped(srch searcher, sol *solution)
	emitted(srch searcher, parent, child *solution)
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

func (ws watchers) emitted(srch searcher, parent, child *solution) {
	for _, w := range ws {
		w.emitted(srch, parent, child)
	}
}

type debugWatcher struct {
	debug  struct {
		expand func(sol, parent *solution)
		before func(sol *solution)
		after  func(sol *solution)
	}
}

func newDebugWatcher(prob *problem) *debugWatcher {
	return &debugWatcher{}
}

func (wat *debugWatcher) emitted(srch searcher, parent, child *solution) {
	if wat.debug.expand != nil {
		wat.debug.expand(child, parent)
	}
}

func (wat *debugWatcher) beforeStep(srch searcher, sol *solution) {
	if wat.debug.before != nil {
		wat.debug.before(sol)
	}
}

func (wat *debugWatcher) stepped(sol *solution) {
	if wat.debug.after != nil {
		wat.debug.after(sol)
	}
}
