package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	dumpProg = flag.Bool("dumpProg", false, "dump the generated search program")
	trace    = flag.Bool("trace", false, "trace results")
	verify   = flag.Bool("verify", false, "generate code for extra verification")
	debug    = flag.Bool("debug", false, "enable debug search watcher")

	prob problem
	srch search
	gg   *goGen
	gen  solutionGen
)

func dump(sol *solution) bool {
	if sol.err == nil {
		gg.logf("=== Solution: %v %s", sol, sol.letterMapping())
	} else if sol.err == errVerifyFailed {
		gg.logf("!!! Fail: %v %s", sol, sol.letterMapping())
	} else if *debug {
		gg.logf("--- Dead end: %v %s", sol, sol.letterMapping())
	} else {
		return false
	}
	for i, soli := range sol.trace {
		gg.logf("... [%v] %v %s", i, soli, soli.letterMapping())
	}
	return false
}

func initSearch(emit emitFunc) {
	emit(newSolution(&prob, gg.steps, emit))
}

func traceFailures() {
	metrics := newMetricWatcher()
	watcher := watchers([]searchWatcher{
		metrics,
		newTraceWatcher(),
	})
	srch.run(100000, initSearch, dump, watcher)
	fmt.Printf("%+v\n", metrics)
}

func debugRun() {
	metrics := newMetricWatcher()
	watcher := watchers([]searchWatcher{
		metrics,
		newTraceWatcher(),
		debugWatcher{
			logf: gg.logf,
		},
	})
	srch.run(100000, initSearch, dump, watcher)
	fmt.Printf("%+v\n", metrics)
}

func findOne() *solution {
	metrics := newMetricWatcher()
	watcher := searchWatcher(metrics)

	if *trace {
		watcher = watchers([]searchWatcher{
			metrics,
			newTraceWatcher(),
		})
	}

	failed := false
	var theSol *solution
	srch.run(
		100000,
		initSearch,
		func(sol *solution) bool {
			if sol.err == errVerifyFailed {
				failed = true
				return false
			}
			if sol.err != nil {
				return false
			}
			if theSol != nil {
				failed = true
				return false
			}
			theSol = sol
			return true
		},
		watcher)
	fmt.Printf("search metrics: %+v\n", metrics)
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
	if err := prob.setup(word1, word2, word3); err != nil {
		log.Fatalf("setup failed: %v", err)
	}

	gg = &goGen{}
	gg.verified = *verify
	gg.debugLabels = *debug || *dumpProg

	if *dumpProg {
		gen = multiGen([]solutionGen{
			&logGen{},
			gg,
			gg.obsAfter(),
		})
	} else {
		gen = gg
	}

	plan(&prob, gen)

	if *dumpProg {
		fmt.Println()
		fmt.Printf("//// Resolved Program Dump\n")
		for i, step := range gg.steps {
			fmt.Printf("%v: %v\n", i, step)
		}
		fmt.Println()
	}

	if *debug {
		debugRun()
		return
	}

	if sol := findOne(); sol != nil {
		gg.logf("found: %v", sol.letterMapping())
		sol.printCheck(gg.logf)
		if sol.trace != nil {
			for i, soli := range sol.trace {
				gg.logf("... [%v] %v %s", i, soli, soli.letterMapping())
			}
		}
	} else {
		traceFailures()
	}
}
