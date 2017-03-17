package main

type machStep interface {
	step(*mach) error
}

type mach struct {
	prog []machStep
	pc   int
}

func newMach(prog []machStep) *mach {
	return &mach{
		prog: prog,
	}
}

func (m *mach) run() error {
	for m.pc < len(m.prog) {
		ms := m.prog[m.pc]
		if err := ms.step(m); err != nil {
			return err
		}
		m.pc++
	}
	return nil
}
