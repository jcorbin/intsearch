package stackvm_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/jcorbin/intsearch/stackvm"
)

func TestVM(t *testing.T) {
	for _, tc := range []vmTestCase{

		{
			name: "23add 5eq",
			code: []interface{}{
				Push(2), Push(3), Add,
				Push(5), Eq,
			},
			stack: []byte{1},
		},

		{
			name: "If 23mul 6eq Then 42 Else 99 End",
			code: []interface{}{
				If, Push(2), Push(3), Mul, Push(6), Eq,
				Then, Push(42),
				Else, Push(99), End,
			},
			stack: []byte{42},
		},

		{
			name: "Each dupdec Loop If dup2mod Then 3mulinc Else 2div End Done",
			code: []interface{}{
				Push(3),
				While, Dup, Dec, Loop,
				If, Dup, Push(2), Mod,
				Then, Push(2), Div,
				Else, Push(3), Mul, Inc, End,
				End,
			},
			stack: []byte{1},
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
	name  string
	code  []interface{}
	err   error
	stack []byte
	heap  []byte
}

func (tc vmTestCase) run(t *testing.T) {
	m, err := Compile(tc.code)
	require.NoError(t, err, "unexpected Compile error")
	if err := m.Run(); tc.err == nil {
		assert.NoError(t, err)
	} else {
		assert.EqualError(t, err, tc.err.Error(), "expected Mach.Run error")
	}
	assert.Equal(t, tc.stack, m.Stack(), "expected final Mach.Stack()")
	assert.Equal(t, tc.heap, m.Heap(), "expected final Mach.Heap()")
}

func (tc vmTestCase) trace(t *testing.T) {
	t.Logf("CODE: %v", tc.code)

	d := dumper{t.Logf}
	m, err := Compile(tc.code)
	require.NoError(t, err, "unexpected Compile error")
	if err := m.Trace(d); tc.err == nil {
		assert.NoError(t, err)
	} else {
		assert.EqualError(t, err, tc.err.Error(), "expected Mach.Run error")
	}
	assert.Equal(t, tc.stack, m.Stack(), "expected final Mach.Stack()")
	assert.Equal(t, tc.heap, m.Heap(), "expected final Mach.Heap()")
}

type dumper struct {
	logf func(string, ...interface{})
}

func (d dumper) Begin(m *Mach) {
	prog := m.Prog()
	w := len(strconv.Itoa(len(prog)))
	d.logf("PROG:")
	for i, op := range prog {
		d.logf("[%*d]: %v", w, i, op)
	}
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

func (d dumper) After(m *Mach, pc int, last, next Op, err error) {
	if err != nil {
		d.logf("--> ERR: % 10v -- heap=%v stack=%v",
			err, m.Heap(), m.Stack())
		return
	}
	d.logf("--> % 3d:            -- heap=%v stack=%v",
		pc, m.Heap(), m.Stack())
}
