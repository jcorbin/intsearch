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
	if err := prob.validate(word1, word2, word3); err != nil {
		return err
	}
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

func (prob *problem) numColumns() int {
	return len(prob.words[2])
}

func (prob *problem) getColumn(k int) [3]byte {
	var (
		cx [3]byte
		w  = len(prob.words[2])
	)
	if i := len(prob.words[0]) - w + k; i >= 0 {
		cx[0] = prob.words[0][i]
	}
	if j := len(prob.words[1]) - w + k; j >= 0 {
		cx[1] = prob.words[1][j]
	}
	cx[2] = prob.words[2][k]
	return cx
}
