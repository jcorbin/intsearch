package main

import (
	"errors"
	"fmt"
	"sort"
)

var errInvalidResultWidth = errors.New("invalid result width, must be equal to or only one greater than the widest argument")

type problem struct {
	words     [3][]byte
	letterSet map[byte]bool
	base      int
	known     map[byte]bool
	gen       solutionGen
}

type solutionGen interface {
	init(prob *problem, desc string)
	fix(prob *problem, c byte, v int)
	initColumn(prob *problem, cx [3]byte, numKnown, numUnknown int)
	solve(prob *problem, neg bool, c byte, c1, c2 byte)
	computeCarry(prob *problem, c1, c2 byte)
	choose(prob *problem, c byte)
	checkFinal(prob *problem, c byte, c1, c2 byte)
	finish(prob *problem)
}

func (prob *problem) plan(word1, word2, word3 string, gen solutionGen) error {
	if err := prob.validate(word1, word2, word3); err != nil {
		return err
	}
	if err := prob.setup(word1, word2, word3); err != nil {
		return err
	}

	prob.gen = gen
	prob.gen.init(prob, "bottom up")

	prob.planBottomUp()
	return nil
}

func (prob *problem) validate(word1, word2, word3 string) error {
	argWidth := len(word1)
	if len(word2) > argWidth {
		argWidth = len(word2)
	}
	resWidthDiff := len(word3) - argWidth
	if resWidthDiff != 0 && resWidthDiff != 1 {
		return errInvalidResultWidth
	}
	return nil
}

func (prob *problem) setup(word1, word2, word3 string) error {
	prob.letterSet = make(map[byte]bool, len(word1)+len(word2)+len(word3))
	for x, word := range []string{word1, word2, word3} {
		prob.words[x] = make([]byte, len(word))
		for i, r := range word {
			c := byte(r)
			prob.words[x][i] = c
			prob.letterSet[c] = true
		}
	}
	prob.base = 10
	if len(prob.letterSet) > 10 {
		return fmt.Errorf("only base 10 problems supported currently")
	}
	prob.known = make(map[byte]bool, len(prob.letterSet))
	return nil
}

func (prob *problem) sortedLetters() []string {
	letters := make([]string, 0, len(prob.letterSet))
	for c := range prob.letterSet {
		letters = append(letters, string(c))
	}
	sort.Strings(letters)
	return letters
}

func (prob *problem) planBottomUp() {
	// for each column from the right
	//   choose letters until 2/3 are known
	//   compute the third (if unknown)

	var (
		cx [3]byte
		ix = [3]int{
			len(prob.words[0]) - 1,
			len(prob.words[1]) - 1,
			len(prob.words[2]) - 1,
		}
	)

	for ix[0] >= 0 || ix[1] >= 0 || ix[2] >= 0 {
		for x, i := range ix {
			if i >= 0 {
				cx[x] = prob.words[x][i]
			} else {
				cx[x] = 0
			}
		}

		prob.solveColumn(cx)

		ix[0]--
		ix[1]--
		ix[2]--
	}

	prob.gen.finish(prob)
}

func (prob *problem) solveColumn(cx [3]byte) {
	numKnown := 0
	numUnknown := 0
	for _, c := range cx {
		if c != 0 {
			if prob.known[c] {
				numKnown++
			}
			if !prob.known[c] {
				numUnknown++
			}
		}
	}

	prob.gen.initColumn(prob, cx, numKnown, numUnknown)

	for x, c := range cx {
		if c != 0 {
			if !prob.known[c] {
				if numUnknown == 1 {
					switch x {
					case 0:
						prob.gen.solve(prob, true, c, cx[2], cx[1])
					case 1:
						prob.gen.solve(prob, true, c, cx[2], cx[0])
					case 2:
						prob.gen.solve(prob, false, c, cx[0], cx[1])
					}
				} else {
					prob.gen.choose(prob, c)
				}
				prob.known[c] = true
				numUnknown--
				numKnown++
			} else if x == 2 && cx[0] == 0 && cx[1] == 0 {
				prob.gen.checkFinal(prob, c, cx[0], cx[1])
			}
		}
	}
	prob.gen.computeCarry(prob, cx[0], cx[1])
}
