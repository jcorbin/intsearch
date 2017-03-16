package main

import "fmt"

type step interface {
}

type problem struct {
	b, n       int
	w1, w2, w3 string
	emit       func(...step)
	cols       []col
	revsym     map[byte]byte
	sym        map[byte]byte
	known      map[byte]struct{}
}

type col [3]byte

func (c col) unknown(known map[byte]struct{}) (int, int) {
	n, i := 0, -1
	for j := range c {
		if c[j] == 0 {
			continue
		}
		if _, have := known[c[j]]; !have {
			n++
			if i < 0 {
				i = j
			}
		}
	}
	return n, i
}

func (p *problem) scan() {
	p.cols = make([]col, len(p.w3))
	p.revsym = make(map[byte]byte, len(p.w1)+len(p.w2)+len(p.w3))
	p.sym = make(map[byte]byte, len(p.w1)+len(p.w2)+len(p.w3))
	for i := 1; i <= len(p.w3); i++ {
		var c col
		for j, w := range []string{p.w1, p.w2, p.w3} {
			if i > len(w) {
				continue
			}
			l := w[len(w)-i]
			s, seen := p.sym[l]
			if !seen {
				p.n++
				s = byte(p.n)
				p.revsym[s] = l
				p.sym[l] = s
			}
			c[j] = s
		}
		p.cols[len(p.cols)-i] = c
	}
}

type rem struct{ s string }

func remf(pat string, args ...interface{}) rem {
	return rem{fmt.Sprintf(pat, args...)}
}

func (r rem) String() string {
	return fmt.Sprintf("-- %s", r.s)
}

func (p *problem) pick(s byte) {
	p.emit(remf(
		"pick %d (%q)",
		s, string(p.revsym[s]),
	))
	p.known[s] = struct{}{}
}

func (p *problem) solve(i int, c col, j int) {
	if i == len(p.cols)-1 {
		p.emit(remf(
			"solve for %q in %q + %q = %q (mod %d)",
			string(p.revsym[c[j]]),
			string(p.revsym[c[0]]),
			string(p.revsym[c[1]]),
			string(p.revsym[c[2]]),
			p.b,
		))
	} else {
		p.emit(remf(
			"solve for %q in carry %q + %q = %q (mod %d)",
			string(p.revsym[c[j]]),
			string(p.revsym[c[0]]),
			string(p.revsym[c[1]]),
			string(p.revsym[c[2]]),
			p.b,
		))
	}
	p.known[c[j]] = struct{}{}
}

func (p *problem) check(i int, c col) {
	if i == len(p.cols)-1 {
		p.emit(remf(
			"check %q + %q == %q (mod %d)",
			string(p.revsym[c[0]]),
			string(p.revsym[c[1]]),
			string(p.revsym[c[2]]),
			p.b,
		))
	} else {
		p.emit(remf(
			"check carry + %q + %q == %q (mod %d)",
			string(p.revsym[c[0]]),
			string(p.revsym[c[1]]),
			string(p.revsym[c[2]]),
			p.b,
		))
	}
}

func (p *problem) computeCarry(i int, c col) {
	if i == len(p.cols)-1 {
		p.emit(remf(
			"compute carry = (%q + %q) / %d",
			string(p.revsym[c[0]]),
			string(p.revsym[c[1]]),
			p.b,
		))
	} else {
		p.emit(remf(
			"compute carry = (carry + %q + %q) / %d",
			string(p.revsym[c[0]]),
			string(p.revsym[c[1]]),
			p.b,
		))
	}
}

func (p *problem) bottomUp() {
	p.known = make(map[byte]struct{}, p.n)
	for i := len(p.cols) - 1; i >= 0; i-- {
		c := p.cols[i]
		n, j := c.unknown(p.known)
		for n > 1 {
			p.pick(c[j])
			n, j = c.unknown(p.known)
		}
		if n == 1 {
			p.solve(i, c, j)
		} else {
			p.check(i, c)
		}
		p.computeCarry(i, c)
	}
}

func plan(w1, w2, w3 string, emit func(...step)) {
	p := problem{
		b:    10,
		w1:   w1,
		w2:   w2,
		w3:   w3,
		emit: emit,
	}
	p.scan()
	p.bottomUp()
}

func printSteps(steps ...step) {
	for i, s := range steps {
		fmt.Printf("% 3d: %v\n", i, s)
	}
}

func main() {
	plan(
		"send", "more", "money",
		printSteps)
}
