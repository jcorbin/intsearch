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
	emit(state)
	fork(*state, int) error
	branch(*state, int) error
}
type step interface {
	run(s *state) error
}

type state struct {
	ctx   context
	pc    int
	prog  []step
	stack int
	heap  int
	mem   [1024]int
}

func (st *state) run() error {
	for st.pc < len(st.prog) {
		op := st.prog[st.pc]
		st.pc++
		if err := op.run(st); err != nil {
			return err
		}
	}
	return nil
}

var (
	errStackOverflow  = errors.New("stack overflow")
	errStackUnderflow = errors.New("stack underflow")
	errOutOfMemory    = errors.New("out of memory")
	errSegFault       = errors.New("memory segmentation fault")
	errUsed           = errors.New("value used")
)

type halt struct{ error }
type push int
type pop struct{}
type alloc int
type dup struct{}
type swap struct{}
type store int
type load int
type loadOffset int
type add struct{}
type sub struct{}
type mul struct{}
type div struct{}
type mod struct{}
type lt struct{}
type lte struct{}
type eq struct{}
type neq struct{}
type gte struct{}
type gt struct{}
type jmp int
type jz int
type jnz int
type fork int
type fz int
type fnz int
type branch int
type bz int
type bnz int

func (op halt) String() string         { return fmt.Sprintf("halt %v", op.error) }
func (op push) String() string         { return fmt.Sprintf("push %d", int(op)) }
func (op pop) String() string          { return fmt.Sprintf("pop") }
func (op alloc) String() string        { return fmt.Sprintf("alloc %d", int(op)) }
func (op dup) String() string          { return fmt.Sprintf("dup") }
func (op swap) String() string         { return fmt.Sprintf("swap") }
func (addr store) String() string      { return fmt.Sprintf("store %d", int(addr)) }
func (addr load) String() string       { return fmt.Sprintf("load %d", int(addr)) }
func (addr loadOffset) String() string { return fmt.Sprintf("loadOffset %d", int(addr)) }
func (op add) String() string          { return fmt.Sprintf("add") }
func (op sub) String() string          { return fmt.Sprintf("sub") }
func (op mul) String() string          { return fmt.Sprintf("mul") }
func (op div) String() string          { return fmt.Sprintf("div") }
func (op mod) String() string          { return fmt.Sprintf("mod") }
func (op lt) String() string           { return fmt.Sprintf("lt") }
func (op lte) String() string          { return fmt.Sprintf("lte") }
func (op eq) String() string           { return fmt.Sprintf("eq") }
func (op neq) String() string          { return fmt.Sprintf("neq") }
func (op gte) String() string          { return fmt.Sprintf("gte") }
func (op gt) String() string           { return fmt.Sprintf("gt") }
func (op jmp) String() string          { return fmt.Sprintf("jmp %d", int(op)) }
func (op jz) String() string           { return fmt.Sprintf("jz %d", int(op)) }
func (op jnz) String() string          { return fmt.Sprintf("jnz %d", int(op)) }
func (op fork) String() string         { return fmt.Sprintf("fork %d", int(op)) }
func (op fz) String() string           { return fmt.Sprintf("fz %d", int(op)) }
func (op fnz) String() string          { return fmt.Sprintf("fnz %d", int(op)) }
func (op branch) String() string       { return fmt.Sprintf("branch %d", int(op)) }
func (op bz) String() string           { return fmt.Sprintf("bz %d", int(op)) }
func (op bnz) String() string          { return fmt.Sprintf("bnz %d", int(op)) }

func (op halt) run(s *state) error {
	s.pc = len(s.prog)
	if op.error != nil {
		s.pc++
	}
	return op.error
}

func (op push) run(s *state) error {
	i := s.stack
	j := i + 1
	if j >= s.heap {
		return errStackOverflow
	}
	s.stack = j
	s.mem[i] = int(op)
	return nil
}

func (op pop) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	s.stack = j
	return nil
}

func (op alloc) run(s *state) error {
	i := s.heap
	j := i - int(op)
	if j <= s.stack {
		return errOutOfMemory
	}
	s.heap = j
	return nil
}

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

func (addr loadOffset) run(s *state) error {
	if int(addr) < s.heap {
		return errSegFault
	}
	i := s.stack
	if i < 0 {
		return errStackUnderflow
	}
	j := int(addr) + s.mem[i]
	if j < s.heap {
		return errSegFault
	}
	s.mem[i] = s.mem[j]
	return nil
}

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

func (op lt) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] < s.mem[i] {
		s.mem[j] = 1
	} else {
		s.mem[j] = 0
	}
	s.stack = j
	return nil
}

func (op lte) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] <= s.mem[i] {
		s.mem[j] = 1
	} else {
		s.mem[j] = 0
	}
	s.stack = j
	return nil
}

func (op eq) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] == s.mem[i] {
		s.mem[j] = 1
	} else {
		s.mem[j] = 0
	}
	s.stack = j
	return nil
}

func (op neq) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] != s.mem[i] {
		s.mem[j] = 1
	} else {
		s.mem[j] = 0
	}
	s.stack = j
	return nil
}

func (op gte) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] >= s.mem[i] {
		s.mem[j] = 1
	} else {
		s.mem[j] = 0
	}
	s.stack = j
	return nil
}

func (op gt) run(s *state) error {
	i := s.stack
	if i < 2 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] > s.mem[i] {
		s.mem[j] = 1
	} else {
		s.mem[j] = 0
	}
	s.stack = j
	return nil
}

func (op jmp) run(s *state) error {
	s.pc += int(op)
	return nil
}

func (op jz) run(s *state) error {
	i := s.stack
	if i < 1 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] == 0 {
		s.pc += int(op)
	}
	s.stack = j
	return nil
}

func (op jnz) run(s *state) error {
	i := s.stack
	if i < 1 {
		return errStackUnderflow
	}
	j := i - 1
	if s.mem[j] != 0 {
		s.pc += int(op)
	}
	s.stack = j
	return nil
}

func (op fork) run(s *state) error {
	return s.ctx.fork(s, int(op))
}

func (op fz) run(s *state) error {
	i := s.stack
	if i < 1 {
		return errStackUnderflow
	}
	j := i - 1
	s.stack = j
	if s.mem[j] == 0 {
		return s.ctx.fork(s, int(op))
	}
	return nil
}

func (op fnz) run(s *state) error {
	i := s.stack
	if i < 1 {
		return errStackUnderflow
	}
	j := i - 1
	s.stack = j
	if s.mem[j] != 0 {
		return s.ctx.fork(s, int(op))
	}
	return nil
}

func (op branch) run(s *state) error {
	return s.ctx.branch(s, int(op))
}

func (op bz) run(s *state) error {
	i := s.stack
	if i < 1 {
		return errStackUnderflow
	}
	j := i - 1
	s.stack = j
	if s.mem[j] == 0 {
		return s.ctx.branch(s, int(op))
	}
	return nil
}

func (op bnz) run(s *state) error {
	i := s.stack
	if i < 1 {
		return errStackUnderflow
	}
	j := i - 1
	s.stack = j
	if s.mem[j] != 0 {
		return s.ctx.branch(s, int(op))
	}
	return nil
}

func plan(w1, w2, w3 string, base int) []step {
	prog := make([]step, 0, 1024)

	// setup
	// Memory Layout:
	// - B-many used values (0 / non-zero)
	// - N-many byte / value pairs

	n := 0
	addr := 1024 - 1 // top of memory
	addr -= base     // used values
	usedAddr := addr

	valueAddr := make(map[byte]int, len(w1)+len(w2)+len(w3))
	for _, w := range []string{w1, w2, w3} {
		for i := range w {
			b := w[i]
			if _, seen := valueAddr[b]; !seen {
				valueAddr[b] = addr
				addr -= 2
				n++
			}
		}
	}

	prog = append(prog, alloc(1024-addr))
	for b, addr := range valueAddr {
		prog = append(prog,
			push(int(b)),
			store(addr-1))
	}

	k := make(known, n)
	cols := words2cols(w1, w2, w3)

	for i, col := range cols {
		// choose until at most one unknown
		n, ci := k.countUnknown(col)
		for n > 1 {
			prog = append(prog,
				push(0),              //
				fork(4),              //
				jmp(6),               // ------\
				push(base-1),         // <---\ |
				lt{},                 //     | |
				bnz(3),               // >-\ | |
				push(1),              //   | | |
				add{},                //   | | |
				jmp(-6),              // >-|-/ |
				dup{},                // <-/---/
				loadOffset(usedAddr), //
				jz(1),
				halt{errUsed},
				dup{},
				store(valueAddr[col[ci]]),
			)
			k.mark(col[ci])
			n, ci = k.countUnknown(col)
		}

		if n == 1 {
			// if we have one unknown, solve for it
			// dup (carry)
			switch ci {
			case 0: // a = c - b - cx
				// neg (carry)
				// load c
				// load b
				// sub
				fmt.Printf("solve(%s)\n", col.equation(0, 2, 1, i, "-", "="))
			case 1: // b = c - a - cx
				// neg (carry)
				// load c
				// load a
				// sub
				fmt.Printf("solve(%s)\n", col.equation(1, 2, 0, i, "-", "="))
			case 2: // c = a + b + cx
				// load a
				// load b
				// add
				fmt.Printf("solve(%s %% %d)\n", col.equation(2, 0, 1, i, "+", "="), base)
			}
			// add (carry)
			// push base
			// mod
			// dup (value)
			// loadOffset usedAddr
			// jz 1
			// halt errUsed
			// dup (value)
			// storeOffset usedAddr
			// store col[ci]
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

	return prog
}

type search struct {
	frontier []state
}

func (s *search) emit(st state) {
	s.frontier = append(s.frontier, st)
}

func (s *search) fork(st *state, off int) error {
	newst := *st
	newst.pc += off
	s.frontier = append(s.frontier, newst)
	return nil
}

func (s *search) branch(st *state, off int) error {
	s.frontier = append(s.frontier, *st)
	st.pc += off
	return nil
}

func (s *search) next() state {
	st := s.frontier[0]
	s.frontier = s.frontier[:copy(s.frontier, s.frontier[1:])]
	return st
}

func initState(ctx context, prog []step) state {
	st := state{prog: prog}
	st.heap = len(st.mem)
	ctx.emit(st)
	return st
}

func main() {
	prog := plan("send", "more", "money", 10)
	for i, op := range prog {
		fmt.Printf("% 3d : %v\n", i, op)
	}

	// var srch search
	// initState(&srch, prog)
	// for len(srch.frontier) > 0 {
	// 	st := srch.next()
	// 	err := st.run()
	// 	if err == nil {
	// 		fmt.Printf("FOUND %+v\n", st)
	// 	}
	// }
}
