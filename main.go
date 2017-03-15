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

func (prob *problem) solveColumn(carry string, a, b, c byte, unk int) {
	if unk == 2 {
		// solve for c
		if carry == "" {
			fmt.Printf("val = %s + %s %% %d\n", string(a), string(b), prob.base)
		} else {
			fmt.Printf("val = %s + %s + %s %% %d\n", string(a), string(b), carry, prob.base)
		}
		fmt.Printf("halt errConflict if used[val] != 0\n")
		fmt.Printf("used[val] = 1\n")
		fmt.Printf("%s = val\n", string(c))
		return
	}

	// solving for b is same as solving for a
	if unk == 1 {
		unk, a, b = 0, b, a
	}

	// solve for a
	if carry == "" {
		fmt.Printf("val = %s - %s %% %d\n", string(c), string(b), prob.base)
	} else {
		fmt.Printf("val = %s - %s - %s %% %d\n", string(c), string(b), carry, prob.base)
	}
	fmt.Printf("halt errConflict if used[val] != 0\n")
	fmt.Printf("mark used[val] = 1\n")
	fmt.Printf("%s = val\n", string(a))
}

func (prob *problem) plan() {
	fmt.Printf("//// setup\n")
	fmt.Printf("// reserve haep space for the used array\n")
	fmt.Printf("alloc %d\n", prob.base)
	fmt.Printf("// reserve haep space for letter values\n")
	fmt.Printf("alloc %d\n", prob.n)

	// TODO: translate all letter access below into values[I]
	// references

	for i, col := range prob.cols {
		fmt.Printf("\n")

		var carry string
		if i > 0 {
			carry = fmt.Sprintf("C%d", i)
		}
		fmt.Printf("//// col[%d]: %v\n", i, col.Equation(carry))

		// until we have a most one unknown, pick a value for the first unknown
		n, first := prob.unknown(col)
		for n > 1 {
			c := col[first]

			fmt.Printf("// pick(%s)\n", string(c))
			fmt.Printf("for 0 <= i < %v {\n", prob.base)
			fmt.Printf("  continue if used[i] != 0\n")
			fmt.Printf("  forkContinue if i < %v\n", prob.base-1)
			fmt.Printf("  %v = i\n", c)
			fmt.Printf("  used[i] = 1\n")
			fmt.Printf("}\n")

			prob.known[c] = struct{}{}
			n, first = prob.unknown(col)
		}

		// if we still have one unknown, solve for it
		if n == 1 {
			fmt.Printf("// solve %s   (mod %d) for %s\n",
				col.Equation(carry), prob.base,
				string(col[first]))
			prob.solveColumn(carry, col[0], col[1], col[2], first)
			prob.known[col[first]] = struct{}{}
		} else {
			// we have no unknows, check
			fmt.Printf("// check col_%d\n", i)
			fmt.Printf("if (%s) %% %d != %s {\n",
				col.RHS(carry), prob.base, string(col[2]))
			fmt.Printf("  halt errCheckFailed")
			fmt.Printf("}\n")
		}

		// compute outgoing carry
		if i < len(prob.cols)-1 {
			j := i + 1
			fmt.Printf("// compute C%d\n", j)
			fmt.Printf("C%d = %s + %s / %d\n",
				j,
				string(col[0]), string(col[1]),
				prob.base)
		}
	}
}

func main() {
	prob := initProblem(10, "send", "more", "money")
	prob.plan()
}
