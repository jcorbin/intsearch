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
		init: func(emit func(*solution)) {
			emit(newSolution(&prob, gg.steps, emit))
		},
	}

	srch.result = func(sol *solution, trace []*solution) {
		if sol.err != nil && sol.err != errVerifyFailed {
			return
		}
		srch.dump(sol, trace)
	}

	// srch.debug = func(before bool, sol *solution) {
	// 	if before {
	// 		fmt.Printf(">>> %v\n", sol)
	// 	} else {
	// 		fmt.Printf("... %v\n", sol)
	// 		if _, ok := sol.steps[sol.stepi-1].(storeStep); ok {
	// 			fmt.Printf("... %s\n", sol.letterMapping())
	// 		} else if _, ok := sol.steps[sol.stepi-1].(forkUntilStep); ok {
	// 			fmt.Printf("... len(frontier) == %v\n", len(srch.frontier))
	// 		}
	// 	}
	// }

	srch.run(100000)
	fmt.Printf("%+v\n", srch.metrics)
}
