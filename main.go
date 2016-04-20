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

	prob   problem
	srch   search
	gg     = goGen{}
	gen    = solutionGen(&gg)
	labels []string
)

func getLabels() []string {
	if labels == nil {
		labels = make([]string, len(gg.steps))
		for label, addr := range gg.labels {
			labels[addr] = label
		}
	}
	return labels
}

func labelFor(sol *solution) string {
	labels := getLabels()
	if sol.stepi >= len(labels) {
		return ""
	}
	label := labels[sol.stepi]
	if len(label) == 0 {
		return ""
	}
	return fmt.Sprintf("  // %s", label)
}

func dump(sol *solution) {
	if sol.err == nil {
		fmt.Printf("=== Solution: %v%s\n=== ", sol, labelFor(sol))
	} else if sol.err == errVerifyFailed {
		fmt.Printf("!!! Fail: %v%s\n!!! ", sol, labelFor(sol))
	} else if *debug {
		fmt.Printf("--- Dead end: %v%s\n--- ", sol, labelFor(sol))
	}
	fmt.Printf("%s\n", sol.letterMapping())
	for i, soli := range sol.trace {
		fmt.Printf("trace[%v] %v %s%s\n", i, soli, soli.letterMapping(), labelFor(soli))
	}
	fmt.Println()
}

func initSearch(emit emitFunc) {
	emit(newSolution(&prob, gg.getSteps(), emit))
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
			labelFor: labelFor,
		},
	})
	srch.run(100000, initSearch, dump, watcher)
	fmt.Printf("%+v\n", metrics)
}

func findOne() *solution {
	metrics := newMetricWatcher()
	failed := false
	var theSol *solution
	srch.run(
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
		},
		metrics)
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

	gg.verified = *verify

	if *dumpProg {
		gen = &multiGen{[]solutionGen{
			&logGen{},
			&gg,
			gg.obsAfter(),
		}}
	}

	if err := prob.setup(word1, word2, word3); err != nil {
		log.Fatalf("setup failed: %v", err)
	}
	plan(&prob, gen)

	if *dumpProg {
		fmt.Println()
		fmt.Printf("//// Resolved Program Dump\n")
		for i, step := range gg.getSteps() {
			fmt.Printf("%v: %v\n", i, step)
		}
		fmt.Println()
	}

	if *debug {
		debugRun()
		return
	}

	if sol := findOne(); sol != nil {
		fmt.Printf("found: %v\n", sol.letterMapping())
		sol.printCheck(func(format string, args ...interface{}) {
			fmt.Printf(format, args...)
			fmt.Println()
		})
	} else {
		traceFailures()
	}
}
