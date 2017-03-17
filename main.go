package main

import "fmt"

func plan(w1, w2, w3 string) (*mach, error) {
	prog := make([]machStep, 0, 512)
	p := newProb(w1, w2, w3)
	if err := p.plan(func(steps ...interface{}) {
		for _, step := range steps {
			fmt.Printf("% 3d: %v\n", len(prog), step)
			if ms, ok := step.(machStep); ok {
				prog = append(prog, ms)
			}
		}
	}); err != nil {
		return nil, err
	}
	return newMach(prog), nil
}

func main() {
	m, err := plan("send", "more", "money")
	if err != nil {
		fmt.Printf("PLAN FAIL: %v\n", err)
		return
	}
	if err := m.run(); err != nil {
		fmt.Printf("RUN FAIL: %v\n", err)
	}
}
