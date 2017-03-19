package stackvm

import "fmt"

// Push pushes a value onto the stack.
type Push byte

type _pop struct{}
type _swap struct{}
type _dup struct{}
type _over struct{}

var (
	// Pop pops the top stack value.
	Pop = _pop{}

	// Swap swaps the top two stack values.
	Swap = _swap{}

	// Dup pushes a copy of top of stack value.
	Dup = _dup{}

	// Over pushes a copy of the second from top stack value.
	Over = _over{}
)

func (op Push) run(m *Mach) error {
	return m.push(byte(op))
}

func (op _pop) run(m *Mach) error {
	_, err := m.pop()
	return err
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

func (op _over) run(m *Mach) error {
	i, err := m.ref(1)
	if err != nil {
		return err
	}
	return m.push(m.mem[i])
}

func (op Push) String() string  { return fmt.Sprintf("push(%d)", byte(op)) }
func (op _pop) String() string  { return "pop" }
func (op _swap) String() string { return "swap" }
func (op _dup) String() string  { return "dup" }
func (op _over) String() string { return "over" }
