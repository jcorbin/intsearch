package stackvm

var (
	// Inc increments the top of stack value.
	Inc = _inc{}

	// Dec increments the top of stack value.
	Dec = _dec{}

	// Add adds the top of stack value to its predecessor.
	Add = _add{}

	// Sub subtracts the top of stack value from its predecessor.
	Sub = _sub{}

	// Mul multiplies the top of stack value into its predecessor.
	Mul = _mul{}

	// Div divides the top of stack value into its predecessor, leaving the
	// quotient.
	Div = _div{}

	// Mod divides the top of stack value into its predecessor, leaving the
	// remainder.
	Mod = _mod{}
)

type _inc struct{}
type _dec struct{}
type _add struct{}
type _sub struct{}
type _mul struct{}
type _div struct{}
type _mod struct{}

func (op _inc) run(m *Mach) error {
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	m.mem[i]++
	return nil
}

func (op _dec) run(m *Mach) error {
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	m.mem[i]--
	return nil
}

func (op _add) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	m.mem[i] += val
	return nil
}

func (op _sub) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	m.mem[i] -= val
	return nil
}

func (op _mul) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	m.mem[i] *= val
	return nil
}

func (op _div) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	m.mem[i] /= val
	return nil
}

func (op _mod) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	m.mem[i] %= val
	return nil
}

func (op _inc) String() string { return "inc" }
func (op _dec) String() string { return "dec" }
func (op _add) String() string { return "add" }
func (op _sub) String() string { return "sub" }
func (op _mul) String() string { return "mul" }
func (op _div) String() string { return "div" }
func (op _mod) String() string { return "mod" }
