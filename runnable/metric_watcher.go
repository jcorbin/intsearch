package runnable

// MetricWatcher implements a set of simple counters for search.
type MetricWatcher struct {
	// TODO: better metric support
	Steps, Emits, MaxFrontierLen int
}

// NewMetricWatcher creates a new metric watcher.
func NewMetricWatcher() *MetricWatcher {
	return &MetricWatcher{}
}

// Emitted increments the Emits counter and MaxFrontierLen gauge.
func (metrics *MetricWatcher) Emitted(srch Searcher, child *Solution) {
	metrics.Emits++
	if fs := srch.FrontierSize(); fs > metrics.MaxFrontierLen {
		metrics.MaxFrontierLen = fs
	}
}

// BeforeStep does nothing.
func (metrics *MetricWatcher) BeforeStep(srch Searcher, sol *Solution) {
}

// Stepped increments the Steps counter.
func (metrics *MetricWatcher) Stepped(srch Searcher, sol *Solution) {
	metrics.Steps++
}
