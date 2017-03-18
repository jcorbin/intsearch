package stackvm

var (
	// Alloc pops a a value from the stack and grows the heap by that much.
	Alloc = _alloc{}

	// Load consumes an offset from the stack, replacing it with the heap
	// value at the given offset.
	Load = _load{}

	// Store consumes an offset and value from the stack, storing that value
	// into the heap at the given offset.
	Store = _store{}
)

type _alloc struct{}
type _load struct{}
type _store struct{}

func (op _alloc) run(m *Mach) error {
	n, err := m.pop()
	if err != nil {
		return err
	}
	return m.alloc(int(n))
}

func (op _load) run(m *Mach) error {
	i, err := m.ref(0)
	if err != nil {
		return err
	}
	val, err := m.load(int(m.mem[i]))
	if err != nil {
		return err
	}
	m.mem[i] = val
	return nil
}

func (op _store) run(m *Mach) error {
	off, err := m.pop()
	if err != nil {
		return err
	}
	val, err := m.pop()
	if err != nil {
		return err
	}
	return m.store(int(off), val)
}

func (op _alloc) String() string { return "alloc" }
func (op _load) String() string  { return "load" }
func (op _store) String() string { return "store" }
