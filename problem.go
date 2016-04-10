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
	known := make(map[rune]bool, len(prob.letterSet))

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
			log.Printf("set carry = 0")
			first = false
		} else {
			log.Printf("set carry = (%v + %v + carry) // %v", string(cx[0]), string(cx[1]), prob.base)
		}

		numKnown := 0
		for x, i := range ix {
			if i >= 0 {
				c := prob.words[x][i]
				cx[x] = c
				if known[c] {
					numKnown++
				}
			} else {
				cx[x] = 0
			}
		}

		log.Printf("column: %v + %v = %v (numKnown: %v)", string(cx[0]), string(cx[1]), string(cx[2]), numKnown)

		for x, c := range cx {
			if c != 0 {
				if !known[c] {
					if numKnown < 2 {
						log.Printf("choose %v (branch by %v)", string(c), prob.base-len(known))
					} else if x == 0 {
						log.Printf("solve %v = %v - %v - carry (mod %v)", string(c), string(cx[2]), string(cx[1]), prob.base)
					} else if x == 1 {
						log.Printf("solve %v = %v - %v - carry (mod %v)", string(c), string(cx[2]), string(cx[0]), prob.base)
					} else if x == 2 {
						log.Printf("solve %v = %v + %v + carry (mod %v)", string(c), string(cx[0]), string(cx[1]), prob.base)
					} else {
						panic("nope solve")
					}
					known[c] = true
					numKnown++
				}
			}
		}

		ix[0]--
		ix[1]--
		ix[2]--
	}
}
