package stackvm

import "fmt"

// Jmp unconditionally jumps by a relative number of operations.
type Jmp int

// Jnz pops a stack value, and jumps if it is non-zero.
type Jnz int

// Jz pops a stack value, and jumps if it is zero.
type Jz int

// Fork
// Fnz
// Fz
// Branch
// Bnz
// Bz

func (op Jmp) run(m *Mach) error {
	return m.jump(int(op))
}

func (op Jnz) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val != 0 {
		return m.jump(int(op))
	}
	return nil
}

func (op Jz) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	if val == 0 {
		m.jump(int(op))
	}
	return nil
}

func (op Jmp) String() string { return fmt.Sprintf("jmp %+d", int(op)) }
func (op Jnz) String() string { return fmt.Sprintf("jnz %+d", int(op)) }
func (op Jz) String() string  { return fmt.Sprintf("jz %+d", int(op)) }
