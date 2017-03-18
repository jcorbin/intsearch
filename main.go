package main

import "fmt"

type annotator interface {
	annotate(ann map[string]int, i int)
}

type resolver interface {
	resolve(ann map[string]int, i int) machStep
}

func plan(w1, w2, w3 string) (*mach, error) {
	prog := make([]machStep, 0, 512)
	p := newProb(w1, w2, w3)
	if err := p.plan(func(steps ...interface{}) {
		type resi struct {
			resolver
			i int
		}
		var (
			res []resi
			ann = make(map[string]int)
		)

		for _, step := range steps {
			i := len(prog)
			switch ts := step.(type) {
			case annotator:
				fmt.Printf(">>> %v\n", step)
				ts.annotate(ann, i)
			case resolver:
				res = append(res, resi{ts, i})
				fmt.Printf("% 3d: TODO %v\n", i, ts)
				prog = append(prog, nil)
			case machStep:
				fmt.Printf("% 3d: %v\n", i, step)
				prog = append(prog, ts)
			default:
				fmt.Printf("-- %v\n", step)
			}
		}

		for _, ri := range res {
			prog[ri.i] = ri.resolver.resolve(ann, ri.i)
			fmt.Printf("RES % 3d: %v\n", ri.i, prog[ri.i])
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
