package main

import "errors"

var (
	errTooManyLetters = errors.New("too many letters")

	// prog errors
	errDead = errors.New("dead")
	errUsed = errors.New("value already used")
)

type col [3]byte

func (c col) unknown(known map[byte]struct{}) (int, int) {
	n, u := 0, -1
	for i, l := range c {
		if _, yes := known[l]; !yes {
			n++
			if u < 0 {
				u = i
			}
		}
	}
	return n, u
}

type prob struct {
	b, n   int
	ws     [3]string
	sym    map[byte]byte
	revsym map[byte]byte
	cols   []col

	// plan state
	known map[byte]struct{}
}

func newProb(w1, w2, w3 string) *prob {
	p := &prob{
		b:  10,
		ws: [3]string{w1, w2, w3},
	}
	return p
}

func (p *prob) scan() error {
	// build enough space so that we don't have to re-allocate
	est := len(p.ws[0]) + len(p.ws[1]) + len(p.ws[2])
	p.sym = make(map[byte]byte, est)
	p.revsym = make(map[byte]byte, est)
	p.cols = make([]col, len(p.ws[2]))

	for i := 0; i < len(p.ws[2]); i++ {
		// build col of letters
		var c col
		if j := len(p.ws[0]) - i - 1; j >= 0 {
			c[0] = p.ws[0][j]
		}
		if j := len(p.ws[1]) - i - 1; j >= 0 {
			c[1] = p.ws[1][j]
		}
		c[2] = p.ws[2][len(p.ws[2])-i-1]

		// symbolicate col
		for j, l := range c {
			s, def := p.sym[l]
			if !def && l != 0 {
				p.n++
				s = byte(p.n)
				p.sym[l] = s
				p.revsym[s] = l
			}
			c[j] = s
		}

		// fill the cols in so that it's natural
		// left-to-right despite our right-to-left loop
		// structure
		p.cols[len(p.cols)-i-1] = c
	}

	// determine base
	p.b = 10 // TODO: other
	if p.n > p.b {
		return errTooManyLetters
	}

	return nil
}

func (p *prob) pick(s byte, emit func(...interface{})) {
	loop := label("loop")
	next := label("return")
	ret := label("return")

	emit(
		push(0),
		loop,

		// fork next if i < b-1
		push(p.b-1), lt,
		next.comeFrom(fnzFrom),

		// return if !used[i] else halt errUsed
		dup, load,
		ret.comeFrom(jzFrom),
		halt(errUsed),

		next,
		push(1), add,
		dup, push(p.b), lt,
		loop.comeFrom(jnzFrom),
		halt(errDead),

		ret,
		dup, push(1), swap, store, // used[i] = 1
		push(p.n+s), store, // value[s] = i
	)
}

// func (p *prob) solve(carry bool, c col, u int, emit func(...interface{}))
// func (p *prob) check(carry bool, c col, emit func(...interface{}))
// func (p *prob) computeCarry(carry bool, c col, emit func(...interface{}))

func (p *prob) bottomUp(emit func(...interface{})) error {
	for i := 1; i <= len(p.cols); i-- {
		c := p.cols[len(p.cols)-i]
		// carry := i > 1

		// pick until we have more than one unknown
		n, u := c.unknown(p.known)
		for n > 1 {
			p.pick(c[u], emit)
			n, u = c.unknown(p.known)
		}

		// // compute the remaining unknown...
		// if n == 1 {
		// 	p.solve(carry, c, u, emit)
		// } else {
		// 	// ...or check if none
		// 	p.check(carry, c, emit)
		// }

		// // compute carry
		// p.computeCarry(carry, c, emit)
	}
	return nil
}

func (p *prob) plan(emit func(...interface{})) error {
	if err := p.scan(); err != nil {
		return err
	}
	p.known = make(map[byte]struct{}, p.n)
	return p.bottomUp(emit)
}
