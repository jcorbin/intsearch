package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	dumpProg = flag.Bool("dumpProg", false, "dump the generated search program")
	verify   = flag.Bool("verify", false, "generate code for extra verification")
	debug    = flag.Bool("debug", false, "enable debug search watcher")

	prob problem
	srch search
	gg   = goGen{}
	gen  = solutionGen(&gg)
)

func initSearch(emit func(*solution)) {
	emit(newSolution(&prob, gg.steps, emit))
}

func traceFailures() {
	traces := newTraceWatcher()
	metrics := newMetricWatcher()
	srch.watcher = watchers([]searchWatcher{
		metrics,
		traces,
	})
	runSearch(
		&srch,
		100000,
		initSearch,
		func(sol *solution) {
			if sol.err == nil {
				fmt.Printf("=== Solution: %v\n=== ", sol)
			} else if sol.err == errVerifyFailed {
				fmt.Printf("!!! Fail: %v\n!!! ", sol)
			}
			fmt.Printf("%s\n", sol.letterMapping())
			trace := traces[sol]
			for i, soli := range trace {
				fmt.Printf("trace[%v] %v %s\n", i, soli, soli.letterMapping())
			}
			fmt.Println()
		})
	fmt.Printf("%+v\n", metrics)
}

func debugRun() {
	traces := newTraceWatcher()
	metrics := newMetricWatcher()
	srch.watcher = watchers([]searchWatcher{
		metrics,
		traces,
		debugWatcher{},
	})
	runSearch(
		&srch,
		100000,
		initSearch,
		func(sol *solution) {
			if sol.err == nil {
				fmt.Printf("=== Solution: %v\n=== ", sol)
			} else if sol.err == errVerifyFailed {
				fmt.Printf("!!! Fail: %v\n!!! ", sol)
			} else {
				fmt.Printf("--- Dead end: %v\n--- ", sol)
			}
			fmt.Printf("%s\n", sol.letterMapping())
			trace := traces[sol]
			for i, soli := range trace {
				fmt.Printf("trace[%v] %v %s\n", i, soli, soli.letterMapping())
			}
			fmt.Println()
		})
	fmt.Printf("%+v\n", metrics)
}

func findOne() *solution {
	failed := false
	var theSol *solution
	runSearch(
		&srch,
		100000,
		initSearch,
		func(sol *solution) {
			if sol.err == errVerifyFailed {
				failed = true
			} else if sol.err == nil {
				if theSol == nil {
					theSol = sol
				} else {
					failed = true
				}
			}
		})
	if !failed {
		return theSol
	}
	return nil
}

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

	gg.verified = *verify

	if *dumpProg {
		gen = &multiGen{[]solutionGen{
			&logGen{},
			&gg,
			gg.obsAfter(),
		}}
	}

	if err := prob.plan(word1, word2, word3, gen); err != nil {
		log.Fatalf("plan failed: %v", err)
	}

	srch.hintFrontier(len(prob.letterSet))

	if *debug {
		debugRun()
		return
	}

	metrics := newMetricWatcher()
	srch.watcher = metrics
	if sol := findOne(); sol != nil {
		fmt.Printf("found: %v\n", sol.letterMapping())
		fmt.Printf("search metrics: %+v\n", metrics)
	} else {
		traceFailures()
	}
}
