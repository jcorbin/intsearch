package main

import (
	"errors"
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
	planName = flag.String("plan", "bottomUp", fmt.Sprintf(
		"which plan strategy to use (%s)",
		strings.Join(planStrategyNames(), ", ")))

	prob word.Problem
	plan word.Plan
)

func logf(format string, args ...interface{}) {
	dec := plan.Decorate(args...)
	if len(dec) > 0 {
		format = fmt.Sprintf("%s  // %s", format, strings.Join(dec, ", "))
	}
	fmt.Printf(format, args...)
	fmt.Println()
}

type dumper struct {
	cont bool
}

func (dmp *dumper) Result(sol word.Solution) bool {
	var mess string
	if err := sol.Check(); err == nil {
		mess = "=== Solution"
	} else if _, is := err.(word.VerifyError); is {
		mess = fmt.Sprintf("!!! %s", err)
	} else if *debug || *dumpAll {
		mess = "--- Dead end"
	}
	if mess == "" {
		return false
	}

	if !dmp.cont {
		dmp.cont = true
	} else {
		fmt.Println()
	}
	sol.Dump(logf)
	// TODO: restore trace print
	return false
}

func traceFailures() {
	// TODO: restore over word.Plan
	// metrics := runnable.NewMetricWatcher()
	// watcher := runnable.Watchers([]runnable.SearchWatcher{
	// 	metrics,
	// 	runnable.NewTraceWatcher(),
	// })
	plan.Run(&dumper{})
	// fmt.Printf("\nsearch metrics: %+v\n", metrics)
}

func debugRun() {
	// TODO: restore over word.Plan
	// metrics := runnable.NewMetricWatcher()
	// watcher := runnable.Watchers([]runnable.SearchWatcher{
	// 	metrics,
	// 	runnable.NewTraceWatcher(),
	// 	runnable.DebugWatcher{
	// 		Logf: logf,
	// 	},
	// })
	plan.Run(&dumper{})
	// fmt.Printf("\nsearch metrics: %+v\n", metrics)
}

var errMoreThanOneSolution = errors.New("more than one solution")

type singleResult struct {
	sol word.Solution
	err error
}

func (sr *singleResult) Result(sol word.Solution) bool {
	if sr.err != nil {
		return false
	}
	if err := sol.Check(); err != nil {
		if _, is := err.(word.VerifyError); is {
			sr.sol = sol
			sr.err = err
		}
		return false
	}
	if sr.sol != nil {
		sr.err = errMoreThanOneSolution
		return false
	}
	sr.sol = sol
	return true
}

func findOne() word.Solution {
	// TODO: restore over word.Plan
	// metrics := runnable.NewMetricWatcher()
	// watcher := runnable.SearchWatcher(metrics)
	// if *trace {
	// 	watcher = runnable.Watchers([]runnable.SearchWatcher{
	// 		metrics,
	// 		runnable.NewTraceWatcher(),
	// 	})
	// }

	var sr singleResult
	plan.Run(&sr)
	// fmt.Printf("search metrics: %+v\n", metrics)
	if sr.err == nil {
		return sr.sol
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

	planf, ok := planStrategies[*planName]
	if !ok {
		log.Fatalf(
			"invalid plan strategy %q, valid choices: %s",
			planName, strings.Join(planStrategyNames(), ", "))
	}

	if err := prob.Setup(word1, word2, word3); err != nil {
		log.Fatalf("setup failed: %v", err)
	}

	annotated := *dumpProg || *trace || *debug
	// - dumping program benefits from annotations
	// - as do program traces
	// - the debug watcher always traces
	gg := runnable.NewStepGen(word.NewPlanProblem(&prob, annotated))

	gen := word.SolutionGen(gg)
	if *dumpProg {
		gen = word.MultiGen([]word.SolutionGen{
			word.NewLogGen(gg.PlanProblem),
			gen,
		})
	}

	plan = planf(gg.PlanProblem, gen, *verify)

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
		// TODO: restore
		// for _, soli := range sol.Trace() {
		// 	soli.Dump(logf)
		// }
	} else {
		logf("found no solutions, re-running with trace")
		traceFailures()
	}
}
