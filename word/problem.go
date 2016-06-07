package word

import (
	"errors"
	"fmt"
	"sort"
)

// ErrInvalidResultWidth indicates that the given words can't work because of
// the result word size.
var ErrInvalidResultWidth = errors.New("invalid result width, must be equal to or only one greater than the widest argument")

type sortBytes []byte

func (sb sortBytes) Len() int           { return len(sb) }
func (sb sortBytes) Less(i, j int) bool { return sb[i] < sb[j] }
func (sb sortBytes) Swap(i, j int)      { sb[i], sb[j] = sb[j], sb[i] }

// Problem is a word triple to solve.
type Problem struct {
	Words   [3][]byte
	Letters map[byte]bool
	Base    int
}

// Setup sets up the problem for three given words, any error returned
// indicates that the words aren't solvable.
func (prob *Problem) Setup(word1, word2, word3 string) error {
	argWidth := len(word1)
	if len(word2) > argWidth {
		argWidth = len(word2)
	}
	resWidthDiff := len(word3) - argWidth
	if resWidthDiff != 0 && resWidthDiff != 1 {
		return ErrInvalidResultWidth
	}

	prob.Letters = make(map[byte]bool, len(word1)+len(word2)+len(word3))
	for x, word := range []string{word1, word2, word3} {
		prob.Words[x] = make([]byte, len(word))
		for i, r := range word {
			c := byte(r)
			prob.Words[x][i] = c
			prob.Letters[c] = true
		}
	}
	prob.Base = 10
	if len(prob.Letters) > 10 {
		return fmt.Errorf("only base 10 problems supported currently")
	}
	return nil
}

// SortedLetters returns the unique letters used in the problem in sorted
// order.
func (prob *Problem) SortedLetters() []byte {
	letters := make([]byte, 0, len(prob.Letters))
	for c := range prob.Letters {
		letters = append(letters, c)
	}
	sort.Sort(sortBytes(letters))
	return letters
}

// NumColumns returns the number of columns in this problem.
func (prob *Problem) NumColumns() int {
	return len(prob.Words[2])
}

// GetColumn returns the characters used for a given column.
func (prob *Problem) GetColumn(k int) [3]byte {
	var (
		cx [3]byte
		w  = len(prob.Words[2])
	)
	if i := len(prob.Words[0]) - w + k; i >= 0 {
		cx[0] = prob.Words[0][i]
	}
	if j := len(prob.Words[1]) - w + k; j >= 0 {
		cx[1] = prob.Words[1][j]
	}
	cx[2] = prob.Words[2][k]
	return cx
}
