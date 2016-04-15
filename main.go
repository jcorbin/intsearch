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

	traces := newTraceWatcher()
	metrics := newMetricWatcher()
	srch := search{
		watcher: watchers([]searchWatcher{
			metrics,
			traces,
			// debugWatcher{},
		}),
	}
	srch.hintFrontier(len(prob.letterSet))

	runSearch(
		&srch,
		100000,
		func(emit func(*solution)) {
			emit(newSolution(&prob, gg.steps, emit))
		},
		func(sol *solution) {
			if sol.err == nil {
				fmt.Printf("=== Solution: %v %v\n", 0, sol)
				fmt.Printf("=== %v %s\n", 0, sol.letterMapping())
			} else if sol.err == errVerifyFailed {
				fmt.Printf("!!! Fail: %v\n", sol)
				fmt.Printf("!!! %s\n", sol.letterMapping())
				trace := traces[sol]
				for i, soli := range trace {
					fmt.Printf("%v %v %s\n", i, soli, soli.letterMapping())
				}
			}
		})
	fmt.Printf("%+v\n", metrics)
}
