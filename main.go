package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jcorbin/intsearch/runnable"
	"github.com/jcorbin/intsearch/word"
)

var planStrategies = map[string]word.PlanFunc{
	"naiveBrute":  word.PlanNaiveBrute,
	"prunedBrute": word.PlanPrunedBrute,
	"bottomUp":    word.PlanBottomUp,
	"topDown":     word.PlanTopDown,
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
	prob  word.Problem
	srch  runnable.Search
	gg    *runnable.StepGen
	gen   word.SolutionGen
)

func logf(format string, args ...interface{}) {
	dec := gg.Decorate(args)
	if len(dec) > 0 {
		format = fmt.Sprintf("%s  // %s", format, strings.Join(dec, ", "))
	}
	fmt.Printf(format, args...)
	fmt.Println()
}

func dump(sol word.Solution) bool {
	var mess string
	if err := sol.Check(); err == nil {
		mess = "=== Solution"
	} else if _, is := err.(runnable.VerifyError); is {
		mess = fmt.Sprintf("!!! %s", err)
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
	sol.Dump(logf)
	for _, soli := range sol.Trace() {
		soli.Dump(logf)
	}

	return false
}

func traceFailures() {
	metrics := runnable.NewMetricWatcher()
	watcher := runnable.Watchers([]runnable.SearchWatcher{
		metrics,
		runnable.NewTraceWatcher(),
	})
	first = true
	srch.Run(gg.SearchInit, func(sol *runnable.Solution) bool {
		return dump(sol)
	}, watcher)
	fmt.Printf("\nsearch metrics: %+v\n", metrics)
}

func debugRun() {
	metrics := runnable.NewMetricWatcher()
	watcher := runnable.Watchers([]runnable.SearchWatcher{
		metrics,
		runnable.NewTraceWatcher(),
		runnable.DebugWatcher{
			Logf: logf,
		},
	})
	first = true
	srch.Run(gg.SearchInit, func(sol *runnable.Solution) bool {
		return dump(sol)
	}, watcher)
	fmt.Printf("\nsearch metrics: %+v\n", metrics)
}

func findOne() word.Solution {
	metrics := runnable.NewMetricWatcher()
	watcher := runnable.SearchWatcher(metrics)

	if *trace {
		watcher = runnable.Watchers([]runnable.SearchWatcher{
			metrics,
			runnable.NewTraceWatcher(),
		})
	}

	failed := false
	var theSol word.Solution
	first = true
	srch.Run(
		gg.SearchInit,
		func(sol *runnable.Solution) bool {
			err := sol.Check()
			if _, is := err.(runnable.VerifyError); is {
				failed = true
				return false
			}
			if err != nil {
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

	if err := prob.Setup(word1, word2, word3); err != nil {
		log.Fatalf("setup failed: %v", err)
	}

	annotated := *dumpProg || *trace || *debug
	// - dumping program benefits from annotations
	// - as do program traces
	// - the debug watcher always traces
	gg = runnable.NewStepGen(word.NewPlanProblem(&prob, annotated))

	if *dumpProg {
		gen = gg.LoggedGen()
	} else {
		gen = gg
	}

	planf(gg.PlanProblem, gen, *verify)

	if *dumpProg {
		fmt.Println()
		fmt.Printf("//// Compiled Program Dump\n")
		for i, step := range gg.Steps() {
			label := gg.LabelAt(i)
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
		logf("found: %v", word.SolutionMapping(sol))
		word.SolutionCheck(sol, logf)
		for _, soli := range sol.Trace() {
			soli.Dump(logf)
		}
	} else {
		logf("found no solutions, re-running with trace")
		traceFailures()
	}
}
