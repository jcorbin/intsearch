package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	flag.Parse()
	word1 := flag.Arg(0)
	if len(word1) == 0 {
		log.Fatalf("missing word1 argument")
	}
	word2 := flag.Arg(1)
	if len(word2) == 0 {
		log.Fatalf("missing word2 argument")
	}
	word3 := flag.Arg(2)
	if len(word3) == 0 {
		log.Fatalf("missing word3 argument")
	}

	var (
		prob problem
		gg   = goGen{
			verified: false,
		}
		gen = multiGen{[]solutionGen{
			&logGen{},
			&gg,
			gg.obsAfter(),
		}}
	)

	if err := prob.plan(word1, word2, word3, &gen); err != nil {
		log.Fatalf("plan failed: %v", err)
	}

	srch := search{
		frontier: make([]*solution, 0, len(prob.letterSet)),
		traces:   make(map[*solution][]*solution, len(prob.letterSet)),
	}

	// srch.debug.emit = func(sol, parent *solution) {
	// 	fmt.Printf("+++ %v %v", len(srch.frontier), sol)
	// 	if parent != nil {
	// 		fmt.Printf(" parent %v @%v", parent.steps[parent.stepi], parent.stepi)
	// 	}
	// 	fmt.Printf("\n")
	// }
	// srch.debug.before = func(sol *solution) {
	// 	fmt.Printf(">>> %v\n", sol)
	// }
	// srch.debug.after = func(sol *solution) {
	// 	fmt.Printf("... %v\n", sol)
	// 	if _, ok := sol.steps[sol.stepi-1].(storeStep); ok {
	// 		fmt.Printf("... %s\n", sol.letterMapping())
	// 	} else if _, ok := sol.steps[sol.stepi-1].(forkUntilStep); ok {
	// 		fmt.Printf("... len(frontier) == %v\n", len(srch.frontier))
	// 	}
	// }

	srch.emit(newSolution(&prob, gg.steps, srch.emit))
	srch.run(100000, func(sol *solution) {
		if sol.err != nil && sol.err != errVerifyFailed {
			return
		}
		trace := srch.traces[sol]
		srch.dump(sol, trace)
	})
	fmt.Printf("%+v\n", srch.metrics)
}
