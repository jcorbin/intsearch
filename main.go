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
		gg   goGen
		gen  = multiGen{[]solutionGen{
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
		result: func(sol *solution, trace []*solution) {
			if sol.err == nil {
				fmt.Println()
				fmt.Println("Solution:")
			} else {
				// fmt.Println()
				// fmt.Printf("Fail: %v\n", err)
				return
			}
			for i, soli := range trace {
				fmt.Printf("%v %v %s\n", i, soli, soli.letterMapping())
			}
			fmt.Printf("=== %v %v\n", 0, sol)
			fmt.Printf("=== %v %s\n", 0, sol.letterMapping())
		},
	}
	srch.run(100000)
}
