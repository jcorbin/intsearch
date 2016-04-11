package main

import (
	"fmt"
	"strings"
)

type logGen struct {
}

func (lg logGen) init(prob *problem, desc string) {
	var w int
	for _, word := range prob.words {
		if len(word) > w {
			w = len(word)
		}
	}
	fmt.Printf("# Problem:\n")
	fmt.Printf("#   %s%v\n", strings.Repeat(" ", w-len(prob.words[0])), string(prob.words[0]))
	fmt.Printf("# + %s%v\n", strings.Repeat(" ", w-len(prob.words[1])), string(prob.words[1]))
	fmt.Printf("# = %s%v\n", strings.Repeat(" ", w-len(prob.words[2])), string(prob.words[2]))
	fmt.Printf("# base: %v\n", prob.base)
	fmt.Printf("# letters: %v\n", prob.sortedLetters())
	fmt.Printf("# method: %s\n", desc)
	fmt.Printf("\n")
	fmt.Printf("// set carry = 0\n")
}

func (lg logGen) interColumn(prob *problem, cx [3]rune) {
	if cx[0] != 0 && cx[1] != 0 {
		fmt.Printf("// set carry = (%v + %v + carry) // %v\n", string(cx[0]), string(cx[1]), prob.base)
	} else if cx[0] != 0 {
		fmt.Printf("// set carry = (%v + carry) // %v\n", string(cx[0]), prob.base)
	} else if cx[1] != 0 {
		fmt.Printf("// set carry = (%v + carry) // %v\n", string(cx[1]), prob.base)
	} else {
		fmt.Printf("// set carry = carry // %v\n", prob.base)
	}
}

func (lg logGen) initColumn(prob *problem, cx [3]rune, numKnown, numUnknown int) {
	if cx[0] != 0 && cx[1] != 0 {
		fmt.Printf("// column: carry + %v + %v = %v\n", string(cx[0]), string(cx[1]), string(cx[2]))
	} else if cx[0] != 0 {
		fmt.Printf("// column: carry + %v = %v\n", string(cx[0]), string(cx[2]))
	} else if cx[1] != 0 {
		fmt.Printf("// column: carry + %v = %v\n", string(cx[1]), string(cx[2]))
	}
}

func (lg logGen) solve(prob *problem, neg bool, c, c1, c2 rune) {
	if c1 != 0 && c2 != 0 {
		if neg {
			fmt.Printf("// solve %v = %v - %v - carry (mod %v)\n", string(c), string(c1), string(c2), prob.base)
		} else {
			fmt.Printf("// solve %v = %v + %v + carry (mod %v)\n", string(c), string(c1), string(c2), prob.base)
		}
	} else if c1 != 0 {
		if neg {
			fmt.Printf("// solve %v = %v - carry (mod %v)\n", string(c), string(c1), prob.base)
		} else {
			fmt.Printf("// solve %v = %v + carry (mod %v)\n", string(c), string(c1), prob.base)
		}
	} else if c2 != 0 {
		if neg {
			fmt.Printf("// solve %v = %v - carry (mod %v)\n", string(c), string(c2), prob.base)
		} else {
			fmt.Printf("// solve %v = %v + carry (mod %v)\n", string(c), string(c2), prob.base)
		}
	} else {
		if neg {
			fmt.Printf("// solve %v = - carry (mod %v)\n", string(c), prob.base)
		} else {
			fmt.Printf("// solve %v = + carry (mod %v)\n", string(c), prob.base)
		}
	}
}

func (lg logGen) choose(prob *problem, c rune) {
	fmt.Printf("// choose %v (branch by %v)\n", string(c), prob.base-len(prob.known))
}

func (lg logGen) checkFinal(prob *problem, c, c1, c2 rune) {
	fmt.Printf("// check %v == carry\n", string(c))
}

func (lg logGen) finish(prob *problem) {
	fmt.Printf("// done\n")
}
