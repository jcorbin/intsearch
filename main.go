package main

import "fmt"

func plan(w1, w2, w3 string) *mach {
	prog := make([]machStep, 0, 512)
	p := newProb(w1, w2, w3)

	p.plan(func(step interface{}) {
		fmt.Printf("% 3d: %v\n", len(prog), step)
		if ms, ok := step.(machStep); ok {
			prog = append(prog, ms)
		}
	})
	return newMach(prog)
}

func main() {
	m := plan("send", "more", "money")
	if err := m.run(); err != nil {
		fmt.Printf("FAIL: %v\n", err)
	}
}
