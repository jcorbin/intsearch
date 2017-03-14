package main

import (
	"errors"
	"fmt"
	"strings"
)

type col [3]byte

func (c col) equation(i, j, k, x int, sign, rel string) string {
	// c[i] REL c[j] SIGN c[k] SIGN cX
	if c[i] == 0 {
		return "InvalidEquation(zero RHS)"
	}
	rhs := c.rhs(j, k)
	if x > 0 {
		rhs = append(rhs, fmt.Sprintf("c%d", x))
	}
	return fmt.Sprintf("%s %s %s",
		string(c[i]), rel, strings.Join(rhs, sign))
}

func (c col) rhs(i, j int) []string {
	rhs := make([]string, 0, 2)
	if c[i] != 0 {
		rhs = append(rhs, string(c[i]))
	}
	if c[j] != 0 {
		rhs = append(rhs, string(c[j]))
	}
	return rhs
}

type known map[byte]struct{}

func (k known) mark(cc byte) { k[cc] = struct{}{} }
func (k known) countUnknown(c col) (int, int) {
	n := 0  // number unknown
	i := -1 // index of first unknown in col
	for j, cc := range c {
		if cc == 0 {
			continue
		}
		if _, have := k[cc]; !have {
			n++
			if i < 0 {
				i = j
			}
		}
	}
	return n, i
}

func words2cols(w1, w2, w3 string) []col {
	r := make([]col, len(w3))
	for i := 0; i < len(w3); i++ {
		var a, b, c byte
		if j := len(w1) - i - 1; j >= 0 {
			a = w1[j]
		}
		if j := len(w2) - i - 1; j >= 0 {
			b = w2[j]
		}
		c = w3[len(w3)-i-1]
		r[i] = [3]byte{a, b, c}
	}
	return r
}

type context interface {
	fork(state)
}
type step interface {
	run(s *state) error
}

type state struct {
	ctx   context
	stack int
	heap  int
	mem   [1024]int
}

var (
	errStackOverflow  = errors.New("stack overflow")
	errStackUnderflow = errors.New("stack underflow")
	errOutOfMemory    = errors.New("out of memory")
	errSegFault       = errors.New("memory segmentation fault")
)

type literal int

func (l literal) run(s *state) error {
	i := s.stack
	j := i + 1
	if j >= s.heap {
		return errStackOverflow
	}
	s.stack = j
	s.mem[i] = int(l)
	return nil
}

type alloc int

func (b alloc) run(s *state) error {
	i := s.heap
	j := i - int(b)
	if j <= s.stack {
		return errOutOfMemory
	}
	s.heap = j
	return nil
}

type dup struct{}
type swap struct{}

func (op dup) run(s *state) error {
	i := s.stack
	j := i + 1
	k := j + 1
	if i <= 0 {
		return errStackUnderflow
	}
	if j >= s.heap {
		return errStackOverflow
	}
	if k >= s.heap {
		return errStackOverflow
	}
	s.stack = j
	s.mem[j] = s.mem[i]
	return nil
}

func (op swap) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	s.mem[i], s.mem[j] = s.mem[j], s.mem[i]
	return nil
}

type store int
type load int

func (addr store) run(s *state) error {
	if int(addr) < s.heap {
		return errSegFault
	}
	i := s.stack
	if i <= 0 {
		return errStackUnderflow
	}
	s.stack = i - 1
	s.mem[addr] = s.mem[i]
	return nil
}

func (addr load) run(s *state) error {
	if int(addr) < s.heap {
		return errSegFault
	}
	i := s.stack
	j := i + 1
	if j >= s.heap {
		return errStackOverflow
	}
	s.stack = j
	s.mem[i] = s.mem[addr]
	return nil
}

type add struct{}
type sub struct{}
type mul struct{}
type div struct{}
type mod struct{}

func (op add) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	s.mem[j] += s.mem[i]
	s.stack = j
	return nil
}

func (op sub) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	s.mem[j] -= s.mem[i]
	s.stack = j
	return nil
}

func (op mul) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	s.mem[j] *= s.mem[i]
	s.stack = j
	return nil
}

func (op div) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	s.mem[j] /= s.mem[i]
	s.stack = j
	return nil
}

func (op mod) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	s.mem[j] %= s.mem[i]
	s.stack = j
	return nil
}

type lt struct{}
type lte struct{}
type eq struct{}
type neq struct{}
type gte struct{}
type gt struct{}

type jz struct{}
type jnz struct{}

type fz struct{}
type fnz struct{}

func plan(w1, w2, w3 string, base int) {
	k := make(known, len(w1)+len(w2)+len(w3))
	cols := words2cols(w1, w2, w3)
	for i, col := range cols {
		// choose until at most one unknown
		n, ci := k.countUnknown(col)
		for n > 1 {
			fmt.Printf("pick(%s)\n", string(col[ci]))
			k.mark(col[ci])
			n, ci = k.countUnknown(col)
		}

		if n == 1 {
			// if we have one unknown, solve for it
			switch ci {
			case 0: // a = c - b - cx
				fmt.Printf("solve(%s)\n", col.equation(0, 2, 1, i, "-", "="))
			case 1: // b = c - a - cx
				fmt.Printf("solve(%s)\n", col.equation(1, 2, 0, i, "-", "="))
			case 2: // c = a + b + cx
				fmt.Printf("solve(%s %% %d)\n", col.equation(2, 0, 1, i, "+", "="), base)
			}
			k.mark(col[ci])
		} else {
			// we know all chars in the column, check
			fmt.Printf("check(%s)\n", col.equation(2, 0, 1, i, "+", "=="))
		}

		// compute the outgoing carry
		if j := i + 1; j < len(cols) {
			fmt.Printf("compute(c%d = %s / %d)\n", j, strings.Join(col.rhs(0, 1), "+"), base)
		}
	}
}

type search struct {
	frontier []state
}

func (s search) fork(st *state) error {
	newst := *st
	if err := literal(0).run(st); err != nil {
		return err
	}
	if err := literal(1).run(&newst); err != nil {
		return err
	}
	s.frontier = append(s.frontier, newst)
	return nil
}

func main() {
	plan("send", "more", "money", 10)
}
