package main

import (
	"errors"
	"fmt"
)

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

var (
	errConflict  = errors.New("computed value conflict")
	errCheckFail = errors.New("column check failed")
)

type rem struct{ s string }
type com struct {
	step
	s string
}
type halt struct{ error }
type push int
type jmp int
type jnz int
type jz int
type fork int
type fnz int
type fz int
type _load struct{}
type _store struct{}
type _dup struct{}
type _swap struct{}
type _add struct{}
type _sub struct{}
type _div struct{}
type _mod struct{}
type _lt struct{}
type _lte struct{}
type _eq struct{}
type _neq struct{}
type _gt struct{}
type _gte struct{}

func (r rem) String() string     { return fmt.Sprintf("-- %s", r.s) }
func (c com) String() string     { return fmt.Sprintf("%v -- %s", c.step, c.s) }
func (op halt) String() string   { return fmt.Sprintf("halt %v", op.error) }
func (op push) String() string   { return fmt.Sprintf("push %d", int(op)) }
func (op jmp) String() string    { return fmt.Sprintf("jmp %d", int(op)) }
func (op jnz) String() string    { return fmt.Sprintf("jnz %d", int(op)) }
func (op jz) String() string     { return fmt.Sprintf("jz %d", int(op)) }
func (op fork) String() string   { return fmt.Sprintf("fork %d", int(op)) }
func (op fnz) String() string    { return fmt.Sprintf("fnz %d", int(op)) }
func (op fz) String() string     { return fmt.Sprintf("fz %d", int(op)) }
func (op _load) String() string  { return "load" }
func (op _store) String() string { return "store" }
func (op _dup) String() string   { return "dup" }
func (op _swap) String() string  { return "swap" }
func (op _add) String() string   { return "add" }
func (op _sub) String() string   { return "sub" }
func (op _div) String() string   { return "div" }
func (op _mod) String() string   { return "mod" }
func (op _lt) String() string    { return "lt" }
func (op _lte) String() string   { return "lte" }
func (op _eq) String() string    { return "eq" }
func (op _neq) String() string   { return "neq" }
func (op _gt) String() string    { return "gt" }
func (op _gte) String() string   { return "gte" }

var (
	load  = _load{}
	store = _store{}
	dup   = _dup{}
	swap  = _swap{}
	add   = _add{}
	sub   = _sub{}
	div   = _div{}
	mod   = _mod{}
	lt    = _lt{}
	lte   = _lte{}
	eq    = _eq{}
	neq   = _neq{}
	gt    = _gt{}
	gte   = _gte{}
)

func remf(pat string, args ...interface{}) rem {
	return rem{fmt.Sprintf(pat, args...)}
}

func comd(s step, pat string, args ...interface{}) com {
	return com{
		step: s,
		s:    fmt.Sprintf(pat, args...),
	}
}

func (p *problem) pick(s byte) {
	p.emit(
		remf(
			"pick %d (%q)",
			s, string(p.revsym[s]),
		),
		push(0),      // ... i=0
		remf("loop"), // ... i
		dup,          // ... i i
		push(p.b-1),  // ... i i b-1
		lt,           // ... i i<b-1
		comd(fnz(1), "fork next"), // ... i
		comd(jmp(6), "continue"),  // ... i
		remf("next"),              // ... i
		push(1),                   // ... i 1
		add,                       // ... i++
		dup,                       // ... i i
		push(p.b),                 // ... i i b
		lt,                        // ... i i<b
		comd(jnz(-11), "loop"), // ... i
		remf("continue"),       // ...
	)
	p.known[s] = struct{}{}
}

func (p *problem) colVal(
	carry bool,
	op step,
	c col,
	ix ...int,
) {
	if carry {
		p.emit(dup) // ... carry carry
	}
	n := 0
	for _, k := range ix {
		if c[k] != 0 {
			n++
			p.emit(push(c[k]), load)
		}
	}
	for k := 0; k < n; k++ {
		p.emit(op)
	}
	if carry {
		p.emit(swap, op)
	}
}

func (p *problem) solve(i int, c col, j int) {
	carry := i < len(p.cols)-1
	if !carry {
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

	// compute the unknown value
	switch j {
	case 0:
		// a in a + b = c
		p.colVal(carry, sub, c, 2, 1)
	case 1:
		// b in a + b = c
		p.colVal(carry, sub, c, 2, 0)
	case 2:
		// c in a + b = c
		p.colVal(carry, add, c, 0, 1)
	}

	p.emit(
		// check that the computed value isn't already used
		dup,  // ... val val
		load, // ... val used[val]
		comd(jz(1), "halt if used"), // ... val
		halt{errConflict},           // ...

		// record the computed value
		push(c[j]), // ... val sym
		push(p.n),  // ... val sym n
		add,        // ... val sym+n
		store,      // ...
	)

	p.known[c[j]] = struct{}{}
}

func (p *problem) check(i int, c col) {
	carry := i < len(p.cols)-1
	if !carry {
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
	p.colVal(carry, add, c, 0, 1) // ... carry => carry val
	p.emit(
		dup,        // ... val val
		push(c[2]), // ... val val sym
		push(p.n),  // ... val val sym n
		add,        // ... val val sym+n
		load,       // ... val value[sym]
		eq,         // ... val==value[sym]
		comd(jnz(1), "halt if =="), // ...
		halt{errCheckFail},         // ...
	)
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

func main() {
	i := 0

	plan(
		"send", "more", "money",
		func(steps ...step) {
			for _, s := range steps {
				if _, ok := s.(rem); ok {
					fmt.Printf("%v\n", s)
					continue
				}
				fmt.Printf("   % 3d: %v\n", i, s)
				i++
			}
		},
	)
}
