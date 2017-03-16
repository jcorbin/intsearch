package main

import "fmt"

type col [3]byte

func (cc col) RHS(carry string) string {
	if carry == "" {
		if cc[0] == 0 && cc[1] == 0 {
			return ""
		}
		if cc[0] == 0 {
			return fmt.Sprintf("%s", string(cc[0]))
		}
		if cc[1] == 0 {
			return fmt.Sprintf("%s", string(cc[1]))
		}
		return fmt.Sprintf("%s + %s", string(cc[0]), string(cc[1]))
	}

	if cc[0] == 0 && cc[1] == 0 {
		return carry
	}
	if cc[0] == 0 {
		return fmt.Sprintf("%s + %s", carry, string(cc[0]))
	}
	if cc[1] == 0 {
		return fmt.Sprintf("%s + %s", carry, string(cc[1]))
	}
	return fmt.Sprintf("%s + %s + %s", carry, string(cc[0]), string(cc[1]))
}

func (cc col) Equation(carry string) string {
	rhs := cc.RHS(carry)
	if rhs == "" {
		return fmt.Sprintf("SHRUG %s", string(cc[2]))
	}
	return fmt.Sprintf("%s = %s", rhs, string(cc[2]))
}

func (cc col) String() string {
	return cc.Equation("C")
}

type problem struct {
	base       int
	n          int
	w1, w2, w3 string
	cols       []col // 0-indexed from the right
	known      map[byte]struct{}
}

func scanWords(w1, w2, w3 string) (map[byte]struct{}, []col) {
	// assume w3 is longest
	// assume w3 is no more than one longer than longest of w2,w1
	// neither w1 nor w2 are empty
	lets := make(map[byte]struct{}, len(w3)+len(w2)+len(w1))
	cols := make([]col, len(w3))
	for i := 0; i < len(w3); i++ {
		var a, b, c byte
		if i < len(w1) {
			a = w1[len(w1)-1-i]
			lets[a] = struct{}{}
		}
		if i < len(w2) {
			b = w2[len(w2)-1-i]
			lets[b] = struct{}{}
		}
		c = w3[len(w3)-1-i]
		cols[i] = col{a, b, c}
		lets[c] = struct{}{}
	}
	return lets, cols
}

func initProblem(base int, w1, w2, w3 string) problem {
	lets, cols := scanWords(w1, w2, w3)
	if len(lets) > base {
		panic("nope") // XXX should error
	}
	prob := problem{
		base:  base,
		n:     len(lets),
		w1:    w1,
		w2:    w2,
		w3:    w3,
		cols:  cols,
		known: make(map[byte]struct{}, len(lets)),
	}
	return prob
}

// return the number, and the first index (0, 1, 2) in col that is unknown
func (prob *problem) unknown(cc col) (int, int) {
	n, first := 0, -1
	for i, c := range cc {
		_, known := prob.known[c]
		if !known && c != 0 {
			n++
			if first < 0 {
				first = i
			}
		}
	}
	return n, first
}

func (prob *problem) pickValue(addr int) {
	fmt.Printf("i = 0\n")                // ... i
	fmt.Printf("loop:\n")                //
	fmt.Printf("dup\n")                  // ... i i
	fmt.Printf("load\n")                 // ... i used[i]
	fmt.Printf("jnz continue\n")         // ... i
	fmt.Printf("dup\n")                  // ... i i
	fmt.Printf("push %d\n", prob.base-1) // ... i i B-1
	fmt.Printf("lt\n")                   // ... i i<B-1
	fmt.Printf("fnz continue\n")         // ... i
	fmt.Printf("dup\n")                  // ... i i
	fmt.Printf("push %d\n", addr)        // ... i i addr
	fmt.Printf("store\n")                // ... i
	fmt.Printf("push 1\n")               // ... i 1
	fmt.Printf("swap\n")                 // ... 1 i
	fmt.Printf("store\n")                // ...
	fmt.Printf("jmp return\n")           // ...
	fmt.Printf("continue:\n")            //
	fmt.Printf("push 1\n")               // ... i 1
	fmt.Printf("add\n")                  // ... ++i
	fmt.Printf("dup\n")                  // ... i i
	fmt.Printf("push %d\n", prob.base)   // ... i i B
	fmt.Printf("lt\n")                   // ... i i<B
	fmt.Printf("jnz loop\n")             // ... i
	fmt.Printf("return:\n")              //
}

func (prob *problem) solveColumn(carry bool, addrs [3]int, unk int) {
	// ... ?carry => ... ?carry
	if carry {
		fmt.Printf("dup\n") // ... carry carry
	}
	var n int
	switch unk {
	case 0:
		n = prob.columnValue(carry, "sub", addrs[1], addrs[2])
	case 1:
		n = prob.columnValue(carry, "sub", addrs[0], addrs[2])
	case 2:
		n = prob.columnValue(carry, "add", addrs[0], addrs[1])
	}
	if n > 0 {
		fmt.Printf("push %d\n", prob.base) // ... val base
		fmt.Printf("mod\n")                // ... val%=base
	}
	fmt.Printf("dup\n")               // ... val val
	fmt.Printf("load\n")              // ... val used[val]
	fmt.Printf("jz +1\n")             // ... val
	fmt.Printf("halt errConflict\n")  //
	fmt.Printf("dup\n")               // ... val val
	fmt.Printf("push 1\n")            // ... val val 1
	fmt.Printf("swap\n")              // ... val 1 val
	fmt.Printf("store\n")             // ... val
	fmt.Printf("push %d\n", addrs[2]) // ... val addrs[2]
	fmt.Printf("store\n")             // ...
}

func (prob *problem) checkColumn(carry string, addrs [3]int) {
	fmt.Printf("halt errCheckFailed if ")
	if prob.opColumn(" + ", carry, addrs[0], addrs[1]) {
		fmt.Printf(" %% %d", prob.base)
	}
	fmt.Printf(" != values[%d]\n", addrs[2])
}

func (prob *problem) computeCarry(carry bool, addrs [3]int) {
	fmt.Printf("C%d = (", j)
	prob.opColumn(" + ", carry, addrs[0], addrs[1])
	fmt.Printf(") / %d\n", prob.base)
}

func (prob *problem) columnValue(carry bool, op string, addrs ...int) int {
	// ?carry => val
	n := 0
	for _, addr := range addrs {
		if addr == 0 {
			continue
		}
		n++
		fmt.Printf("push %d\n", addr) // ... addr
		fmt.Printf("load\n")          // ... arg=*addr
	}
	for i := 1; i < n; i++ {
		fmt.Printf("%s\n", op) // ... ?carry arg arg => ... valO=arg
	}
	if carry && n > 0 {
		fmt.Printf("swap\n")   // ... val carry
		fmt.Printf("%s\n", op) // ... valO=carry
	}
	return n
}

func (prob *problem) opColumn(op, carry string, addrs ...int) bool {
	open := false
	if carry != "" {
		fmt.Printf("carry")
		open = true
	}
	any := false
	for _, addr := range addrs {
		if addr == 0 {
			continue
		}
		any = true
		if open {
			fmt.Printf(op)
		}
		fmt.Printf("values[%d]", addr)
		open = true
	}
	return any
}

func (prob *problem) plan() {
	fmt.Printf("-- setup\n")
	fmt.Printf("- reserve haep space for the used array\n")
	fmt.Printf("alloc %d\n", prob.base)
	fmt.Printf("- reserve haep space for letter values\n")
	fmt.Printf("alloc %d\n", prob.n)

	valueAddrs := make(map[byte]int, prob.n)

	for i, col := range prob.cols {

		// assign letter value memory addresses on a first-encountered basis
		var addrs [3]int
		for i, c := range col {
			if c == 0 {
				continue
			}
			if addr, defined := valueAddrs[c]; defined {
				addrs[i] = addr
			} else {
				addrs[i] = prob.base + len(valueAddrs)
				valueAddrs[c] = addrs[i]
			}
		}

		fmt.Printf("\n")

		var carry string
		if i > 0 {
			carry = fmt.Sprintf("C%d", i)
		}
		fmt.Printf("-- col[%d]: %v\n", i, col.Equation(carry))
		fmt.Printf("--   values at %v\n", addrs)

		// until we have a most one unknown, pick a value for the first unknown
		n, first := prob.unknown(col)
		for n > 1 {
			c := col[first]
			fmt.Printf("- pick(%s)\n", string(c))
			prob.pickValue(addrs[first])
			prob.known[c] = struct{}{}
			n, first = prob.unknown(col)
		}

		// if we still have one unknown, solve for it
		if n == 1 {
			fmt.Printf("- solve %s   (mod %d) for %s\n",
				col.Equation(carry), prob.base,
				string(col[first]))
			prob.solveColumn(carry != "", addrs, first)
			prob.known[col[first]] = struct{}{}
		} else {
			// we have no unknows, check
			fmt.Printf("- check col_%d\n", i)
			prob.checkColumn(carry, addrs)
		}

		// compute outgoing carry
		if i < len(prob.cols)-1 {
			j := i + 1
			fmt.Printf("- compute C%d\n", j)
			prob.computeCarry(carry, addrs)
		}
	}
}

func main() {
	prob := initProblem(10, "send", "more", "money")
	prob.plan()
}
