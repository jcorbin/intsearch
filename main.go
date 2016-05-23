package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"
)

var planStrategies = map[string]planFunc{
	"naiveBrute":  planNaiveBrute,
	"prunedBrute": planPrunedBrute,
	"bottomUp":    planBottomUp,
	"topDown":     planTopDown,
}

func planStrategyNames() []string {
	var names []string
	for name := range planStrategies {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

var (
	dumpProg = flag.Bool("dumpProg", false, "dump the generated search program")
	dumpAll  = flag.Bool("dumpAll", false, "dump all solutions")
	trace    = flag.Bool("trace", false, "trace results")
	verify   = flag.Bool("verify", false, "generate code for extra verification")
	debug    = flag.Bool("debug", false, "enable debug search watcher")
	plan     = flag.String("plan", "bottomUp", fmt.Sprintf(
		"which plan strategy to use (%s)",
		strings.Join(planStrategyNames(), ", ")))

	first bool
	prob  problem
	srch  search
	gg    *goGen
	gen   solutionGen
)

func logf(format string, args ...interface{}) {
	dec := gg.decorate(args)
	if len(dec) > 0 {
		format = fmt.Sprintf("%s  // %s", format, strings.Join(dec, ", "))
	}
	fmt.Printf(format, args...)
	fmt.Println()
}

func dump(sol *solution) bool {
	var mess string
	if sol.err == nil {
		mess = "=== Solution"
	} else if isVerifyError(sol.err) {
		mess = fmt.Sprintf("!!! %s", sol.err)
	} else if *debug || *dumpAll {
		mess = "--- Dead end"
	}
	if mess == "" {
		return false
	}

	if first {
		first = false
	} else {
		fmt.Println()
	}
	logf("%s: %v %s", mess, sol, sol.letterMapping())
	printTrace(sol)
	return false
}

func printTrace(sol *solution) {
	for i, soli := range sol.trace {
		var step solutionStep
		if soli.stepi < len(soli.steps) {
			step = soli.steps[soli.stepi]
		}

		trail := "//"
		if label := gg.labelFor(soli.stepi); len(label) > 0 {
			trail = fmt.Sprintf("// %-40s %s", soli.letterMapping(), label)
		} else if mapping := soli.letterMapping(); len(mapping) > 0 {
			trail = fmt.Sprintf("// %s", mapping)
		}

		fmt.Printf("... %3v: ra:%-3v rb:%-3v rc:%-3v done:%v err:%v -- @%-3v %-20v  %s\n",
			i,
			soli.ra, soli.rb, soli.rc,
			soli.done, soli.err,
			soli.stepi, step,
			trail)
	}
}

func traceFailures() {
	metrics := newMetricWatcher()
	watcher := watchers([]searchWatcher{
		metrics,
		newTraceWatcher(),
	})
	first = true
	srch.run(gg.searchInit, dump, watcher)
	fmt.Printf("\nsearch metrics: %+v\n", metrics)
}

func debugRun() {
	metrics := newMetricWatcher()
	watcher := watchers([]searchWatcher{
		metrics,
		newTraceWatcher(),
		debugWatcher{
			logf: logf,
		},
	})
	first = true
	srch.run(gg.searchInit, dump, watcher)
	fmt.Printf("\nsearch metrics: %+v\n", metrics)
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
	first = true
	srch.run(
		gg.searchInit,
		func(sol *solution) bool {
			if isVerifyError(sol.err) {
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

	planf, ok := planStrategies[*plan]
	if !ok {
		log.Fatalf(
			"invalid plan strategy %q, valid choices: %s",
			plan, strings.Join(planStrategyNames(), ", "))
	}

	if err := prob.setup(word1, word2, word3); err != nil {
		log.Fatalf("setup failed: %v", err)
	}

	gg = newGoGen(newPlanProblem(&prob), true)

	if *dumpProg {
		gen = gg.loggedGen()
	} else {
		gen = gg
	}

	planf(gg.planProblem, gen, *verify)

	if *dumpProg {
		fmt.Println()
		fmt.Printf("//// Compiled Program Dump\n")
		for i, step := range gg.steps {
			label := gg.labelFor(i)
			if label == "" {
				fmt.Printf("%v: %v\n", i, step)
			} else {
				fmt.Printf("%v: %v  // %s\n", i, step, label)
			}
		}
		fmt.Println()
	}

	if *debug {
		debugRun()
		return
	}

	if sol := findOne(); sol != nil {
		logf("found: %v", sol.letterMapping())
		sol.printCheck(logf)
		printTrace(sol)
	} else {
		logf("found no solutions, re-running with trace")
		traceFailures()
	}
}
