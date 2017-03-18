package stackvm_test

import (
	"fmt"
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
			name: "if 23mul 6eq then 42 else 99 end",
			code: []interface{}{
				If, Push(2), Push(3), Mul, Push(6), Eq,
				Then, Push(42),
				Else, Push(99), End,
			},
			stack: []byte{43},
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
	fmt.Printf("CODE: %v\n", tc.code)

	var d dumper
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

type dumper struct{}

func (d dumper) Begin(m *Mach) {
	prog := m.Prog()
	w := len(strconv.Itoa(len(prog)))
	fmt.Printf("PROG:\n")
	for i, op := range prog {
		fmt.Printf("[%*d]: %v\n", w, i, op)
	}
	fmt.Printf("BEGIN @%v\n", m.PC())
}
func (d dumper) End(m *Mach, err error) {
	if err != nil {
		fmt.Printf("END ERR @%v\n", m.PC())
		return
	}
	fmt.Printf("END @%v\n", m.PC())
}

func (d dumper) Before(m *Mach, pc int, op Op) {
	fmt.Printf(">>> % 3d: % 10v -- heap=%v stack=%v\n",
		pc, op, m.Heap(), m.Stack())
}

func (d dumper) After(m *Mach, pc int, last, next Op, err error) {
	if err != nil {
		fmt.Printf("--> ERR: % 10v -- heap=%v stack=%v\n",
			err, m.Heap(), m.Stack())
		return
	}
	fmt.Printf("--> % 3d:            -- heap=%v stack=%v\n",
		pc, m.Heap(), m.Stack())
}
