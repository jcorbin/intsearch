package main

type search struct {
	frontier []*solution
	traces   map[*solution][]*solution
	init     func(func(*solution))
	result   func(*solution, []*solution)
}

func (srch *search) emit(sol *solution) {
	// fmt.Printf("+++ %v %v\n", len(srch.frontier), sol)
	srch.frontier = append(srch.frontier, sol)
	parent := srch.frontier[0]
	if srch.traces != nil {
		var trace []*solution
		if parent != nil {
			trace = append(trace, srch.traces[parent]...)
		}
		srch.traces[sol] = trace
	}
}

func (srch *search) step(sol *solution) {
	// fmt.Printf(">>> %v %v\n", 0, sol)
	if srch.traces != nil {
		srch.traces[sol] = append(srch.traces[sol], sol.copy())
	}
	sol.step()
	if sol.err != nil {
		sol.stepi--
		// fmt.Printf("!!! %v %v\n", 0, sol)
		if srch.traces != nil {
			delete(srch.traces, sol)
		}
		srch.frontier = srch.frontier[1:]
	} else if sol.done {
		var trace []*solution
		if srch.traces != nil {
			trace = srch.traces[sol]
			delete(srch.traces, sol)
		}
		srch.result(sol, trace)
		srch.frontier = srch.frontier[1:]
		// } else {
		// 	fmt.Printf("... %v %v\n", 0, sol)
		// 	if _, ok := sol.steps[sol.stepi-1].(storeStep); ok {
		// 		fmt.Printf("... %v %s\n", 0, sol.letterMapping())
		// 	}
	}
}

func (srch *search) run(maxSteps int) bool {
	srch.init(srch.emit)
	counter := 0
	for len(srch.frontier) > 0 {
		counter++
		if counter > maxSteps {
			return false
		}
		srch.step(srch.frontier[0])
	}
	return true
}
