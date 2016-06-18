package runnable_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jcorbin/intsearch/internal"
	"github.com/jcorbin/intsearch/runnable"
	"github.com/jcorbin/intsearch/word"
)

func TestStepGen_prunedBrute(t *testing.T) {
	runStepGenTest(t, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkStepGenPlan_prunedBrute(b *testing.B) {
	benchStepGenPlan(b, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkStepGenRun_prunedBrute(b *testing.B) {
	benchStepGenRun(b, word.PlanPrunedBrute, "send", "more", "money")
}

func TestStepGen_bottomUp(t *testing.T) {
	runStepGenTest(t, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkStepGenPlan_bottomUp(b *testing.B) {
	benchStepGenPlan(b, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkStepGenRun_bottomUp(b *testing.B) {
	benchStepGenRun(b, word.PlanBottomUp, "send", "more", "money")
}

func TestStepGen_topDown(t *testing.T) {
	runStepGenTest(t, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkStepGenPlan_topDown(b *testing.B) {
	benchStepGenPlan(b, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkStepGenRun_topDown(b *testing.B) {
	benchStepGenRun(b, word.PlanTopDown, "send", "more", "money")
}

func runStepGenTest(t *testing.T, planf word.PlanFunc, w1, w2, w3 string) {
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

	gg := runnable.NewStepGen(word.NewPlanProblem(&prob, false))
	plan = planf(gg.PlanProblem, gg, true)

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
		gg = runnable.NewStepGen(word.NewPlanProblem(&prob, true))
		plan = planf(gg.PlanProblem, word.MultiGen([]word.SolutionGen{
			word.NewLogGen(gg.PlanProblem),
			gg,
		}), true)
		plan.Run(word.NewDebugWatcher(logf))
	}
}

func benchStepGenPlan(b *testing.B, planf word.PlanFunc, w1, w2, w3 string) {
	var prob word.Problem
	if err := prob.Setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}
	for n := 0; n < b.N; n++ {
		gg := runnable.NewStepGen(word.NewPlanProblem(&prob, false))
		planf(gg.PlanProblem, gg, false)
	}
}

func benchStepGenRun(b *testing.B, planf word.PlanFunc, w1, w2, w3 string) {
	var (
		prob word.Problem
		plan word.Plan
	)
	if err := prob.Setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	gg := runnable.NewStepGen(word.NewPlanProblem(&prob, false))
	plan = planf(gg.PlanProblem, gg, false)

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
