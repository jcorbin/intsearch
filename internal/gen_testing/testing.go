package gen_testing

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jcorbin/intsearch/internal"
	"github.com/jcorbin/intsearch/word"
)

type genFunc func(*word.PlanProblem) word.SolutionGen

// RunGenTest tests a SolutionGen against a particular planner.
func RunGenTest(
	t *testing.T,
	genf genFunc,
	planf word.PlanFunc,
	w1, w2, w3 string,
) {
	var (
		prob word.Problem
		plan word.Plan
	)
	if err := prob.Setup(w1, w2, w3); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	logf := func(format string, args ...interface{}) {
		dec := plan.Decorate(args)
		if len(dec) > 0 {
			format = fmt.Sprintf("%s  // %s", format, strings.Join(dec, ", "))
		}
		t.Logf(format, args...)
	}

	gen := genf(word.NewPlanProblem(&prob, false))
	plan = planf(gen, true)

	numGood := 0
	plan.Run(word.ResultFunc(func(sol word.Solution) bool {
		err := sol.Check()
		if _, is := err.(word.VerifyError); is {
			sol.Dump(internal.PrefixedF(logf, "!!! invalid solution found:", "..."))
			t.Fail()
		} else if sol.Check() == nil {
			numGood++
		}
		return false
	}))
	if numGood == 0 {
		t.Logf("didn't find any solution")
		t.Fail()
	} else if numGood > 1 {
		t.Logf("found too many solutions: %v", numGood)
		t.Fail()
	}

	if t.Failed() {
		gen = genf(word.NewPlanProblem(&prob, false))
		plan = planf(word.MultiGen([]word.SolutionGen{
			word.NewLogGen(gen.Problem()),
			gen,
		}), true)
		plan.Run(word.NewDebugWatcher(logf))
	}
}

// BenchGenPlan benchmarks a planner against a particular SolutionGen.
func BenchGenPlan(
	b *testing.B,
	genf genFunc,
	planf word.PlanFunc,
	w1, w2, w3 string,
) {
	var prob word.Problem
	if err := prob.Setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}
	for n := 0; n < b.N; n++ {
		gen := genf(word.NewPlanProblem(&prob, false))
		planf(gen, false)
	}
}

// BenchGenRun benchmarks a Plan generated by a particular SolutionGen.
func BenchGenRun(
	b *testing.B,
	genf genFunc,
	planf word.PlanFunc,
	w1, w2, w3 string,
) {
	var (
		prob word.Problem
	)
	if err := prob.Setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	gen := genf(word.NewPlanProblem(&prob, false))
	plan := planf(gen, false)

	for n := 0; n < b.N; n++ {
		numGood := 0
		plan.Run(word.ResultFunc(func(sol word.Solution) bool {
			if sol.Check() == nil {
				numGood++
			}
			return false
		}))
		if numGood == 0 {
			b.Fatalf("didn't find any solution")
		} else if numGood > 1 {
			b.Fatalf("found too many solutions: %v", numGood)
		}
		if b.Failed() {
			break
		}
	}
}
