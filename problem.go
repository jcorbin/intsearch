package main

import (
	"errors"
	"fmt"
	"log"
	"sort"
)

var errInvalidResultWidth = errors.New("invalid result width, must be equal to or only one greater than the widest argument")

type problem struct {
	words     [3][]rune
	letterSet map[rune]bool
	base      int
	known     map[rune]bool
}

func (prob *problem) plan(word1, word2, word3 string) error {
	if err := prob.validate(word1, word2, word3); err != nil {
		return err
	}
	if err := prob.setup(word1, word2, word3); err != nil {
		return err
	}

	// log.Printf("letters: %v", prob.sortedLetters())

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
	prob.letterSet = make(map[rune]bool, len(word1)+len(word2)+len(word3))
	for x, word := range []string{word1, word2, word3} {
		prob.words[x] = make([]rune, len(word))
		for i, c := range word {
			prob.words[x][i] = c
			prob.letterSet[c] = true
		}
	}
	prob.base = 10
	if len(prob.letterSet) > 10 {
		return fmt.Errorf("only base 10 problems supported currently")
	}
	prob.known = make(map[rune]bool, len(prob.letterSet))
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
		cx    [3]rune
		first = true
		ix    = [3]int{
			len(prob.words[0]) - 1,
			len(prob.words[1]) - 1,
			len(prob.words[2]) - 1,
		}
	)

	for ix[0] >= 0 || ix[1] >= 0 || ix[2] >= 0 {
		if first {
			log.Printf("// set carry = 0")
			first = false
		} else {
			log.Printf("// set carry = (%v + %v + carry) // %v", string(cx[0]), string(cx[1]), prob.base)
		}

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
}

func (prob *problem) solveColumn(cx [3]rune) {
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

	if cx[0] != 0 && cx[1] != 0 {
		log.Printf("// column: carry + %v + %v = %v", string(cx[0]), string(cx[1]), string(cx[2]))
	} else if cx[0] != 0 {
		log.Printf("// column: carry + %v = %v", string(cx[0]), string(cx[2]))
	} else if cx[1] != 0 {
		log.Printf("// column: carry + %v = %v", string(cx[1]), string(cx[2]))
	}

	for x, c := range cx {
		if c != 0 {
			if !prob.known[c] {
				if numUnknown == 1 {
					var (
						c1, c2 rune
						neg    bool
					)

					switch x {
					case 0:
						c1, c2, neg = cx[2], cx[1], true
					case 1:
						c1, c2, neg = cx[2], cx[0], true
					case 2:
						c1, c2, neg = cx[0], cx[1], false
					}

					if neg {
						log.Printf("// solve %v = %v - %v - carry (mod %v)", string(c), string(c1), string(c2), prob.base)
					} else {
						log.Printf("// solve %v = %v + %v + carry (mod %v)", string(c), string(c1), string(c2), prob.base)
					}
				} else {
					log.Printf("// choose %v (branch by %v)", string(c), prob.base-len(prob.known))
				}
				prob.known[c] = true
				numUnknown--
				numKnown++
			} else if x == 2 && cx[0] == 0 && cx[1] == 0 {
				log.Printf("// check %v == carry", string(c))
			}
		}
	}
}
