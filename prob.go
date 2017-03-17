package main

type prob struct {
	b, n       int
	w1, w2, w3 string
}

func newProb(w1, w2, w3 string) *prob {
	p := &prob{
		b:  10,
		w1: w1,
		w2: w2,
		w3: w3,
	}
	return p
}

func (p *prob) plan(emit func(interface{})) {
}
