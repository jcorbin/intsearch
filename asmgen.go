package main

import (
	"fmt"
)

type asmGen struct {
}

func (ag *asmGen) init(prob *problem, desc string) {
	fmt.Printf("loadc 0\n")
}

func (ag *asmGen) fix(prob *problem, c rune, v int) {
	fmt.Printf("store RESULT + %v, %v\n", string(c), v)
}

func (ag *asmGen) interColumn(prob *problem, cx [3]rune) {
	fmt.Printf("loada c\n")
	if cx[0] != 0 {
		fmt.Printf("loadb RESULT + %v\n", string(cx[0]))
		fmt.Printf("add a, b\n")
	}
	if cx[1] != 0 {
		fmt.Printf("loadb RESULT + %v\n", string(cx[1]))
		fmt.Printf("add a, b\n")
	}
	fmt.Printf("floordiv %v\n", prob.base)
	fmt.Printf("loadc a\n")
}

func (ag *asmGen) initColumn(prob *problem, cx [3]rune, numKnown, numUnknown int) {
}

func (ag *asmGen) solve(prob *problem, neg bool, c, c1, c2 rune) {
	if c1 == 0 && c2 == 0 {
		fmt.Printf("loada c\n")
	} else if c1 != 0 && c2 != 0 {
		fmt.Printf("loada %v\n", string(c1))
		fmt.Printf("loadb %v\n", string(c2))
		fmt.Printf("add a, b\n")
		fmt.Printf("add a, c\n")
	} else if c1 != 0 {
		fmt.Printf("loada %v\n", string(c1))
		fmt.Printf("add a, c\n")
	} else {
		fmt.Printf("loada %v\n", string(c2))
		fmt.Printf("add a, c\n")
	}
	if neg {
		fmt.Printf("negate\n")
	}
	fmt.Printf("mod %v\n", prob.base)
	fmt.Printf("store RESULT + %v, $a\n", string(c))
	fmt.Printf("store USED + $a, 1\n")
}

func (ag *asmGen) choose(prob *problem, c rune) {
	fmt.Printf("choose %v\n", string(c))
	fmt.Printf("loadb 0\n")
	fmt.Printf(":loop\n")
	fmt.Printf("loada b\n")
	fmt.Printf("inc b\n")

	fmt.Printf("loadb USED + $a\n")
	fmt.Printf("jz :maybe_fork\n")

	fmt.Printf("lt b, %v\n", prob.base)
	fmt.Printf("jnz +1\n")
	fmt.Printf("exit 1\n")
	fmt.Printf("jmp :loop\n")

	fmt.Printf(":maybe_fork\n")
	fmt.Printf("lt b, %v\n", prob.base)
	fmt.Printf("jnz :use\n")
	fmt.Printf("fork\n")
	fmt.Printf("jnz :use\n")
	fmt.Printf("jmp :loop\n")

	fmt.Printf(":use\n")
	fmt.Printf("loada b\n")
	fmt.Printf("loadb 1\n")
	fmt.Printf("sub a, 1\n")
	fmt.Printf("store RESULT + %v, $a\n", string(c))
	fmt.Printf("store USED + $a, 1\n")
}

func (ag *asmGen) checkFinal(prob *problem, c, c1, c2 rune) {
	fmt.Printf("loada %v\n", string(c))
	fmt.Printf("eq a, c\n")
	fmt.Printf("jz +1\n")
	fmt.Printf("exit 1\n")
}

func (ag *asmGen) finish(prob *problem) {
	fmt.Printf("exit 0\n")
}
