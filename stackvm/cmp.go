package stackvm

var (
	// Lt consumes a value from the stack, and then replaces the next value
	// with 1 if it is less than the consumed value, 0 otherwise.
	Lt = _lt{}

	// LtE consumes a value from the stack, and then replaces the next value
	// with 1 if it is less than or equal to the consumed value, 0 otherwise.
	LtE = _lte{}

	// Eq consumes a value from the stack, and then replaces the next value
	// with 1 if it is equal to the consumed value, 0 otherwise.
	Eq = _eq{}

	// Neq consumes a value from the stack, and then replaces the next value
	// with 1 if it is not equal to the consumed value, 0 otherwise.
	Neq = _neq{}

	// GtE consumes a value from the stack, and then replaces the next value
	// with 1 if it is greater than or equal to the consumed value, 0 otherwise.
	GtE = _gte{}

	// Gt consumes a value from the stack, and then replaces the next value
	// with 1 if it is greater than the consumed value, 0 otherwise.
	Gt = _gt{}
)

type _lt struct{}
type _lte struct{}
type _eq struct{}
type _neq struct{}
type _gte struct{}
type _gt struct{}

func (op _lt) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	if m.mem[i] < val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}

func (op _lte) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	if m.mem[i] <= val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}

func (op _eq) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	if m.mem[i] == val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}

func (op _neq) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	if m.mem[i] != val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}

func (op _gte) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	if m.mem[i] >= val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}

func (op _gt) run(m *Mach) error {
	val, err := m.pop()
	if err != nil {
		return err
	}
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	if m.mem[i] > val {
		m.mem[i] = 1
	} else {
		m.mem[i] = 0
	}
	return nil
}

func (op _lt) String() string  { return "lt" }
func (op _lte) String() string { return "lte" }
func (op _eq) String() string  { return "eq" }
func (op _neq) String() string { return "neq" }
func (op _gte) String() string { return "gte" }
func (op _gt) String() string  { return "gt" }
