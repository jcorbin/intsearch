package word

import (
	"fmt"
	"sync"
)

type planProblemPool struct {
	sync.Pool
}

func (pp *planProblemPool) Get() *PlanProblem {
	if item := pp.Pool.Get(); item != nil {
		return item.(*PlanProblem)
	}
	return nil
}

func (pp *planProblemPool) Put(prob *PlanProblem) {
	pp.Pool.Put(prob)
}

// PlanProblem tracks solution planning state for a problem.
type PlanProblem struct {
	Problem
	pool         planProblemPool
	Annotated    bool
	Columns      []Column
	LetCols      map[byte][]*Column
	Known        map[byte]bool
	FixedLetters map[byte]int
	FixedValues  []bool
	remap        map[*Column]*Column
}

// SolutionGen is the interface implemented by plan execution engines.
type SolutionGen interface {
	Logf(string, ...interface{}) error
	Init(desc string)
	Fork(prob *PlanProblem, name, alt, cont string) SolutionGen
	Fix(c byte, v int)
	ComputeSum(col *Column)
	ComputeFirstSummand(col *Column)
	ComputeSecondSummand(col *Column)
	ChooseRange(c byte, min, max int)
	CheckColumn(col *Column, err error)
	Check(err error)
	Finish()
	Verify()
	Finalize()
}

// NewPlanProblem creates a new planning state for a given problem.
func NewPlanProblem(p *Problem, annotated bool) *PlanProblem {
	C := p.NumColumns()
	N := len(p.Letters)
	prob := &PlanProblem{
		Problem:      *p,
		Annotated:    annotated,
		Columns:      make([]Column, C),
		LetCols:      make(map[byte][]*Column, N),
		Known:        make(map[byte]bool, N),
		FixedLetters: make(map[byte]int, N),
		FixedValues:  make([]bool, p.Base),
	}
	var last *Column
	for i := 0; i < C; i++ {
		col := &prob.Columns[i]
		col.I = i
		if i == 0 {
			col.Carry = CarryZero
		} else {
			col.Carry = CarryUnknown
		}
		if last != nil {
			last.Prior = col
		}
		col.Chars = prob.GetColumn(i)
		a, b, c := col.Chars[0], col.Chars[1], col.Chars[2]
		if a != 0 {
			col.Have++
			col.Unknown++
			prob.LetCols[a] = append(prob.LetCols[a], col)
		}
		if b != 0 {
			col.Have++
			if b != a {
				col.Unknown++
				prob.LetCols[b] = append(prob.LetCols[b], col)
			}
		}
		if c != 0 {
			col.Have++
			if c != b && c != a {
				col.Unknown++
				prob.LetCols[c] = append(prob.LetCols[c], col)
			}
		}
		last = col
	}
	return prob
}

func (prob *PlanProblem) copy() *PlanProblem {
	C := prob.NumColumns()
	N := len(prob.Letters)

	other := prob.pool.Get()
	if other == nil {
		other = &PlanProblem{
			pool:      prob.pool,
			remap:     prob.remap,
			Problem:   prob.Problem,
			Annotated: prob.Annotated,
		}
		other.Columns = make([]Column, C)
		other.LetCols = make(map[byte][]*Column, N)
		other.Known = make(map[byte]bool, N)
		other.FixedLetters = make(map[byte]int, N)
		other.FixedValues = append([]bool(nil), prob.FixedValues...)
	} else {
		if len(other.Columns) != C {
			other.Columns = make([]Column, C)
		}
		if len(other.FixedValues) != len(prob.FixedValues) {
			other.FixedValues = append([]bool(nil), prob.FixedValues...)
		}
		for l := range other.FixedLetters {
			delete(other.FixedLetters, l)
		}
		for c := range other.Known {
			delete(other.Known, c)
		}
	}

	for l, v := range prob.FixedLetters {
		other.FixedLetters[l] = v
	}

	for c, k := range prob.Known {
		other.Known[c] = k
	}

	if prob.remap == nil {
		prob.remap = make(map[*Column]*Column, len(prob.Columns))
		other.remap = prob.remap
	}
	remap := prob.remap

	var last *Column
	for i := 0; i < C; i++ {
		other.Columns[i] = prob.Columns[i]
		col := &other.Columns[i]
		remap[&prob.Columns[i]] = col
		if last != nil {
			last.Prior = col
		}
		last = col
	}

	for c, cols := range prob.LetCols {
		otherCols, _ := other.LetCols[c]
		if otherCols == nil {
			otherCols = make([]*Column, len(cols))
		}
		for i, col := range cols {
			otherCols[i] = remap[col]
		}
		other.LetCols[c] = otherCols
	}

	return other
}

func (prob *PlanProblem) markKnown(c byte) {
	prob.Known[c] = true
	for _, col := range prob.LetCols[c] {
		col.Unknown--
		col.Known++
	}
}

// AssumeCarrySolveColumn generates two futures, using SolutionGen.Fork, to
// solve the column by assuming its prior carry to be 0 or 1.
func (prob *PlanProblem) AssumeCarrySolveColumn(
	gen SolutionGen, col *Column,
	andThen func(*PlanProblem, SolutionGen, *Column) bool,
) bool {
	var label, altLabel, contLabel string
	if prob.Annotated {
		label = col.Label()
		gen.Logf("assumeCarrySolveColumn: %s", label)
		label = fmt.Sprintf("assumeCarry(%s)", label)
	}

	altProb := prob.copy()
	altCol := &altProb.Columns[col.I]

	altCol.Prior.Carry = CarryZero
	col.Prior.Carry = CarryOne
	if prob.Annotated {
		altLabel = fmt.Sprintf("assumeCarry(%s)", altCol.Label())
		contLabel = fmt.Sprintf("assumeCarry(%s)", col.Label())
	}
	altGen := gen.Fork(altProb, label, altLabel, contLabel)

	if !altProb.MaySolveColumn(altGen, altCol) ||
		!andThen(altProb, altGen, altCol.Prior) {
		// TODO: needs to be able to cancel the fork, leaving only the cont
		// path below
		panic("alt pruning unimplemented")
	}
	prob.pool.Put(altProb)
	altProb, altGen, altCol = nil, nil, nil

	if !prob.MaySolveColumn(gen, col) ||
		!andThen(prob, gen, col.Prior) {
		// TODO: needs to be able replace the fork with just the generated alt
		// steps, or return false if the alt failed as well
		panic("alt swapping unimplemented")
	}

	return true
}

// CheckColumn generates a check to solve the column, if it not already solved;
// uses SolutionGen.CheckColumn.
func (prob *PlanProblem) CheckColumn(gen SolutionGen, col *Column) bool {
	if !col.Solved {
		gen.CheckColumn(col, nil)
		col.Solved = true
		col.Carry = CarryComputed
	}
	return true
}

// MaySolveColumn generates code to solve the column and returns true if
// possible, false otherwise:
// - if the column is already solved, noop
// - if there are no unknowns, use CheckColumn
// - if there is only one unknown, attempt a heuristic which may be able to fix
//   the remaining character
// - finally, if the prior column's carry is not unknown, then choose and
//   compute characters directly
func (prob *PlanProblem) MaySolveColumn(gen SolutionGen, col *Column) bool {
	if col.Solved {
		if col.Unknown != 0 {
			panic("invalid column solved state")
		}
		return true
	}

	if col.Unknown == 0 {
		return prob.CheckColumn(gen, col)
	}

	if col.Have == 1 {
		if prob.solveSingularColumn(gen, col) {
			return true
		}
	}

	return prob.solveColumnFromPrior(gen, col)
}

// Fix generates code to fix the value of a character; uses SolutionGen.Fix.
func (prob *PlanProblem) Fix(gen SolutionGen, c byte, v int) {
	gen.Fix(c, v)
	prob.FixedLetters[c] = v
	prob.FixedValues[v] = true
	// TODO: consider inlining markKnown and unifying the for range letCols loops
	prob.markKnown(c)
	for _, col := range prob.LetCols[c] {
		col.Fixed++
	}
}

func (prob *PlanProblem) solveSingularColumn(gen SolutionGen, col *Column) bool {
	if col.Have != 1 || col.Unknown != 1 {
		return false
	}

	if c := col.Chars[2]; col.I == 0 && c != 0 {
		// carry + _ + _ = c --> c == carry --> c = carry = 1
		if col.Prior == nil {
			panic("invalid final column: has no prior")
		}
		prob.Fix(gen, c, 1)
		col.Solved = true
		col.Prior.Carry = CarryOne
		return true
	}

	return false
}

func (prob *PlanProblem) fixRange(min, max int, c byte) (int, int) {
	if min == 0 && (prob.FixedValues[0] ||
		c == prob.Words[0][0] ||
		c == prob.Words[1][0] ||
		c == prob.Words[2][0]) {
		min = 1
	}
	for max > 0 && prob.FixedValues[max] {
		max--
	}
	for min <= max && prob.FixedValues[min] {
		min++
	}
	if min > max {
		panic("no choices possible")
	}
	return min, max
}

func (prob *PlanProblem) chooseOne(gen SolutionGen, col *Column) bool {
	return prob.chooseFirst(gen, col)
}

func (prob *PlanProblem) chooseFirst(gen SolutionGen, col *Column) bool {
	for _, cc := range col.Chars {
		if cc == 0 || prob.Known[cc] {
			continue
		}
		min, max := prob.fixRange(0, prob.Base-1, cc)
		if min == max {
			prob.Fix(gen, cc, min)
			return true
		}
		prob.ChooseRange(gen, cc, min, max)
		return true
	}
	return false
}

func (prob *PlanProblem) chooseBest(gen SolutionGen, col *Column) bool {
	var min, max, N [3]int

	i := -1
	for j, cc := range col.Chars {
		if cc == 0 || prob.Known[cc] {
			continue
		}
		min[j], max[j] = prob.fixRange(0, prob.Base-1, cc)
		if min[j] == max[j] {
			prob.Fix(gen, cc, min[j])
			return true
		}
		N[j] = max[j] - min[j]
		if i < 0 || N[j] < N[i] {
			i = j
		}
	}
	if i < 0 {
		return false
	}

	prob.ChooseRange(gen, col.Chars[i], min[i], max[i])
	return true
}

// ChooseRange generates code to try a range of values for a given character;
// uses SolutionGen.ChooseRange.
func (prob *PlanProblem) ChooseRange(gen SolutionGen, c byte, min, max int) {
	gen.ChooseRange(c, min, max)
	prob.markKnown(c)
}

func (prob *PlanProblem) solveColumnFromPrior(gen SolutionGen, col *Column) bool {
	if col.Prior != nil && col.Prior.Carry == CarryUnknown {
		// unknown prior carry is a case for assumeCarrySolveColumn
		return false
	}

	if col.Unknown == 0 {
		// this is checkColumn's job
		return false
	}

	if prob.Annotated {
		gen.Logf("solveFromPrior: %s", col.Label())
	}

	for u := col.Unknown; u > 1; u = col.Unknown {
		if !prob.chooseOne(gen, col) {
			break
		}
	}
	if col.Unknown > 1 {
		// chooseOne was unable to figure it out
		return false
	}

	if c := col.Chars[0]; c != 0 && !prob.Known[c] {
		gen.ComputeFirstSummand(col)
		col.Carry = CarryComputed
		prob.markKnown(c)
	} else if c := col.Chars[1]; c != 0 && !prob.Known[c] {
		gen.ComputeSecondSummand(col)
		col.Carry = CarryComputed
		prob.markKnown(c)
	} else if c := col.Chars[2]; c != 0 && !prob.Known[c] {
		gen.ComputeSum(col)
		col.Carry = CarryComputed
		prob.markKnown(c)
	} else {
		panic("invalid solveColumnFromPrior state")
	}

	col.Solved = true
	return true
}
