package main

type metricWatcher struct {
	// TODO: better metric support
	Steps, Emits, MaxFrontierLen int
}

func newMetricWatcher() *metricWatcher {
	return &metricWatcher{}
}

func (metrics *metricWatcher) dump(sol *solution) {
}

func (metrics *metricWatcher) emitted(srch searcher, parent, child *solution) {
	metrics.Emits++
	if fs := srch.frontierSize(); fs > metrics.MaxFrontierLen {
		metrics.MaxFrontierLen = fs
	}
}

func (metrics *metricWatcher) beforeStep(srch searcher, sol *solution) {
	metrics.Steps++
}

func (metrics *metricWatcher) stepped(srch searcher, sol *solution) {
}
