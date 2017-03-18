package stackvm

import "fmt"

// Push pushes a value onto the stack.
type Push byte

type _swap struct{}
type _dup struct{}

var (
	// Swap swaps the top two stack values.
	Swap = _swap{}

	// Dup pushes a copy of top of stack value.
	Dup = _dup{}
)

func (op Push) run(m *Mach) error {
	return m.push(byte(op))
}

func (op _swap) run(m *Mach) error {
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	j, err := m.ref(1)
	if err != nil {
		return err
	}
	m.mem[i], m.mem[j] = m.mem[j], m.mem[i]
	return nil
}

func (op _dup) run(m *Mach) error {
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	return m.push(m.mem[i])
}

func (op Push) String() string  { return fmt.Sprintf("push(%d)", byte(op)) }
func (op _swap) String() string { return "swap" }
func (op _dup) String() string  { return "dup" }
