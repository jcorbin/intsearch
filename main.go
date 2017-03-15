package main

import "fmt"

type col [3]byte

func (cc col) Equation(carry string) string {
	if carry == "" {
		if cc[0] == 0 && cc[1] == 0 {
			return fmt.Sprintf("SHRUG %s", string(cc[2]))
		} else if cc[0] == 0 {
			return fmt.Sprintf("%s = %s",
				string(cc[0]), string(cc[2]))
		} else if cc[1] == 0 {
			return fmt.Sprintf("%s = %s",
				string(cc[1]), string(cc[2]))
		}
		return fmt.Sprintf("%s + %s = %s",
			string(cc[0]), string(cc[1]), string(cc[2]))
	}

	if cc[0] == 0 && cc[1] == 0 {
		return fmt.Sprintf("%s = %s",
			carry,
			string(cc[2]))
	} else if cc[0] == 0 {
		return fmt.Sprintf("%s + %s = %s",
			carry,
			string(cc[0]), string(cc[2]))
	} else if cc[1] == 0 {
		return fmt.Sprintf("%s + %s = %s",
			carry,
			string(cc[1]), string(cc[2]))
	}
	return fmt.Sprintf("%s + %s + %s = %s",
		carry,
		string(cc[0]), string(cc[1]), string(cc[2]))
}

func (cc col) String() string {
	return cc.Equation("C")
}

type problem struct {
	base       int
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
		}
		if i < len(w2) {
			b = w2[len(w2)-1-i]
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
		if !known {
			n++
			if first < 0 {
				first = i
			}
		}
	}
	return n, first
}

func (prob *problem) plan() {
	for i, col := range prob.cols {
		var carry string
		if i > 0 {
			carry = fmt.Sprintf("C%d", i)
		}

		// until we have a most one unknown, pick a value for the first unknown
		n, first := prob.unknown(col)
		for n > 1 {
			c := col[first]

			fmt.Printf("// pick(%s)\n", string(c))
			fmt.Printf("for 0 <= i < %v\n", prob.base)
			fmt.Printf("  continue if i is used\n")
			fmt.Printf("  forkContinue if i < %v\n", prob.base-1)
			fmt.Printf("  assign %v = i\n", c)
			fmt.Printf("  mark i used\n")

			prob.known[c] = struct{}{}
			n, first = prob.unknown(col)
		}

		// if we still have one unknown, solve for it
		if n == 1 {
			// TODO: actual formula
			fmt.Printf("solve( %s   (mod %d) for %s )\n",
				col.Equation(carry), prob.base,
				string(col[first]))
			prob.known[col[first]] = struct{}{}
		} else {
			// we have no unknows, check
			fmt.Printf("check( %s   (mod %d) )\n",
				col.Equation(carry), prob.base)
		}

		// compute outgoing carry
		if i < len(prob.cols)-1 {
			fmt.Printf("compute(C%d = %s + %s / %d)\n",
				i+1,
				string(col[0]), string(col[1]),
				prob.base)
		}
	}
}

func main() {
	prob := initProblem(10, "send", "more", "money")
	prob.plan()
}
