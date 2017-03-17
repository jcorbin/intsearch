package main

import "errors"

var (
	errTooManyLetters = errors.New("too many letters")

	// prog errors
	errDead  = errors.New("dead")
	errUsed  = errors.New("value already used")
	errCheck = errors.New("column check failed")
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

	emit(
		push(0),
		loop,

		push(p.b-1), lt, next.comeFrom(fnzFrom), // fork next if i < b-1

		dup, load, hnz(errUsed), // halt if used[i]

		next,
		push(1), add,
		dup, push(p.b), lt,
		loop.comeFrom(jnzFrom),
		halt(errDead),

		dup, push(1), swap, store, // used[i] = 1
		push(p.n+int(s)), store, // value[s] = i
	)
}

func (p *prob) solve(carry bool, c col, u int, emit func(...interface{})) {
	// determine op and values under op:
	//   0 -> solve for a in a + b = c
	//        compute a = c - b - carry % B
	//   1 -> solve for b in a + b = c
	//        compute b = c - a - carry % B
	//   2 -> solve for c in a + b = c
	//        compute c = a + b + carry % B
	var (
		op interface{}
		ix [2]int
	)
	switch u {
	case 0:
		op = sub
		ix = [2]int{2, 1}
	case 1:
		op = sub
		ix = [2]int{2, 0}
	case 2:
		op = add
		ix = [2]int{0, 1}
	}

	// emit steps for the computation determined by ix and op
	p.colVal(emit, carry, op, c[ix[0]], c[ix[1]])

	emit(
		push(p.b), mod,
		dup, load, hnz(errUsed), // halt if used[i]
		dup, push(1), swap, store, // used[i] = 1
		push(p.n+int(c[u])), store, // value[col[u]] = i
	)
}

func (p *prob) check(carry bool, c col, emit func(...interface{})) {
	p.colVal(emit, carry, add, c[0], c[1])
	emit(
		push(p.b), mod,
		push(p.n+int(c[2])), load,
		eq, hz(errCheck),
	)
}

func (p *prob) computeCarry(carry bool, c col, emit func(...interface{})) {
	p.colVal(emit, false, add, c[0], c[1])
	if carry {
		emit(add)
	}
	emit(push(p.b), div)
}

func (p *prob) colVal(emit func(...interface{}), carry bool, op interface{}, syms ...byte) {
	n := 0
	if carry {
		emit(dup)
		n++
	}
	for _, s := range syms {
		if s != 0 {
			emit(push(p.n), push(s), add, load) // value[s]
			n++
		}
	}
	for i := 1; i < n; i++ {
		emit(op)
	}
}

func (p *prob) bottomUp(emit func(...interface{})) error {
	for i := 1; i <= len(p.cols); i-- {
		c := p.cols[len(p.cols)-i]
		carry := i > 1

		// pick until we have more than one unknown
		n, u := c.unknown(p.known)
		for n > 1 {
			p.pick(c[u], emit)
			n, u = c.unknown(p.known)
		}

		// compute the remaining unknown...
		if n == 1 {
			p.solve(carry, c, u, emit)
		} else {
			// ...or check if none
			p.check(carry, c, emit)
		}

		// compute carry
		p.computeCarry(carry, c, emit)
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
