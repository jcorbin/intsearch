package stackvm_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/jcorbin/intsearch/stackvm"
)

func TestVM_Run(t *testing.T) {
	for _, tc := range []vmTestCase{

		{
			name: "23add 5eq",
			code: []interface{}{
				Push(2), Push(3), Add,
				Push(5), Eq,
			},
			results: []vmTestResult{
				{stack: []byte{1}},
			},
		},

		{
			name: "If 23mul 6eq Then 42 Else 99 End",
			code: []interface{}{
				If, Push(2), Push(3), Mul, Push(6), Eq,
				Then, Push(42),
				Else, Push(99), End,
			},
			results: []vmTestResult{
				{stack: []byte{42}},
			},
		},

		{
			name: "collatz(3)",
			code: []interface{}{
				Push(3),
				Each, Dup, Dec, Do,
				If, Dup, Push(2), Mod,
				Then, Push(2), Div,
				Else, Push(3), Mul, Inc, End,
				End,
			},
			results: []vmTestResult{
				{stack: []byte{1}},
			},
		},

		{
			name: "much collatz(x) for 0 <= x <= 9",
			code: []interface{}{
				Push(5), Alloc,
				Push(0),
				Push(0), Much, Dup, Push(9), Lt, Then, Inc, End,
				Each, Dup, Dec, Do,
				If, Dup, Push(2), Mod,
				Then, Push(3), Mul, Inc,
				Else, Push(2), Div, End,
				Swap, Inc, Over, Over, Store, Swap,
				End,
			},

			results: []vmTestResult{

				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
				{stack: []byte{1}},
			},
		},
	} {

		t.Run(tc.name, func(t *testing.T) {
			var tt testing.T
			tc.run(&tt)
			if tt.Failed() {
				tc.trace(t)
			}
		})
	}
}

type vmTestCase struct {
	name    string
	code    []interface{}
	err     error
	results []vmTestResult
}

type vmTestResult struct {
	err   error
	stack []byte
	heap  []byte
}

func (tc vmTestCase) run(t *testing.T) {
	m := tc.makeMachine(t)
	if err := m.Run(); tc.err == nil {
		assert.NoError(t, err)
	} else {
		assert.EqualError(t, err, tc.err.Error(), "expected Mach.Run error")
	}
	if len(tc.results) == 1 {
		tc.results[0].check(t, m)
	}
}

func (tc vmTestCase) trace(t *testing.T) {
	// t.Logf("CODE: %v", tc.code)

	t.Logf("CODE")
	w := len(strconv.Itoa(len(tc.code)))
	for i, code := range tc.code {
		t.Logf("[%*d]: %v", w, i, code)
	}

	m := tc.makeMachine(t)

	t.Logf("PROG:")
	prog := m.Prog()
	w = len(strconv.Itoa(len(prog)))
	for i, op := range prog {
		t.Logf("[%*d]: %v", w, i, op)
	}

	if err := m.Trace(dumper{t.Logf}); tc.err == nil {
		assert.NoError(t, err)
	} else {
		assert.EqualError(t, err, tc.err.Error(), "expected Mach.Run error")
	}
	if len(tc.results) == 1 {
		tc.results[0].check(t, m)
	}
}

var errTooMuch = errors.New("too many results")

func (tc vmTestCase) makeMachine(t *testing.T) *Mach {
	m, err := Compile(tc.code)
	require.NoError(t, err, "unexpected Compile error")
	if len(tc.results) > 1 {
		i := 0
		m.Handle(len(tc.results)-1, HandleFunc(func(n *Mach) error {
			if i < len(tc.results) {
				tc.results[i].check(t, n)
				i++
				return nil
			}
			return errTooMuch
		}))
	}
	return m
}

func (res vmTestResult) check(t *testing.T, m *Mach) {
	if err := m.Err(); res.err == nil {
		assert.NoError(t, err, "unexpected result error")
	} else {
		assert.EqualError(t, err, res.err.Error(), "expected result error")
	}
	assert.Equal(t, res.stack, m.Stack(), "expected final Mach.Stack()")
	assert.Equal(t, res.heap, m.Heap(), "expected final Mach.Heap()")
}

type dumper struct {
	logf func(string, ...interface{})
}

func (d dumper) Begin(m *Mach) {
	d.logf("BEGIN @%v", m.PC())
}
func (d dumper) End(m *Mach, err error) {
	if err != nil {
		d.logf("END ERR @%v", m.PC())
		return
	}
	d.logf("END @%v", m.PC())
}

func (d dumper) Before(m *Mach, pc int, op Op) {
	d.logf(">>> % 3d: % 10v -- heap=%v stack=%v",
		pc, op, m.Heap(), m.Stack())
}

func (d dumper) Fork(m, n *Mach, pc int, next Op) {
	d.logf("==> % 3d: % 10v -- heap=%v stack=%v",
		pc, next, n.Heap(), n.Stack())
}

func (d dumper) After(m *Mach, pc int, last, next Op, err error) {
	if err != nil {
		d.logf("--> ERR: % 10v -- heap=%v stack=%v",
			err, m.Heap(), m.Stack())
		return
	}
	d.logf("--> % 3d:            -- heap=%v stack=%v",
		pc, m.Heap(), m.Stack())
}
