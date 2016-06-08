package runnable_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jcorbin/intsearch/runnable"
	"github.com/jcorbin/intsearch/word"
)

type planFunc func(*word.PlanProblem, word.SolutionGen, bool)

func TestGogen_prunedBrute(t *testing.T) {
	runGogenTest(t, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkGogenPlan_prunedBrute(b *testing.B) {
	benchGogenPlan(b, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkGogenRun_prunedBrute(b *testing.B) {
	benchGogenRun(b, word.PlanPrunedBrute, "send", "more", "money")
}

func TestGogen_bottomUp(t *testing.T) {
	runGogenTest(t, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkGogenPlan_bottomUp(b *testing.B) {
	benchGogenPlan(b, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkGogenRun_bottomUp(b *testing.B) {
	benchGogenRun(b, word.PlanBottomUp, "send", "more", "money")
}

func TestGogen_topDown(t *testing.T) {
	runGogenTest(t, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkGogenPlan_topDown(b *testing.B) {
	benchGogenPlan(b, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkGogenRun_topDown(b *testing.B) {
	benchGogenRun(b, word.PlanTopDown, "send", "more", "money")
}

func runGogenTest(t *testing.T, planf word.PlanFunc, w1, w2, w3 string) {
	var prob word.Problem
	if err := prob.Setup(w1, w2, w3); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	var gg *runnable.StepGen

	logf := func(format string, args ...interface{}) {
		dec := gg.Decorate(args)
		if len(dec) > 0 {
			format = fmt.Sprintf("%s  // %s", format, strings.Join(dec, ", "))
		}
		t.Logf(format, args...)
	}

	gg = runnable.NewStepGen(word.NewPlanProblem(&prob, false))
	planf(gg.PlanProblem, gg, true)

	numGood := 0

	resultFunc := func(sol *runnable.Solution) bool {
		if _, is := sol.Err().(runnable.VerifyError); is {
			logf("!!! invalid solution found: %v %s", sol, sol.LetterMapping())
			for i, soli := range sol.Trace() {
				logf("trace[%v]: %v %s", i, soli, soli.LetterMapping())
			}
			t.Fail()
		} else if sol.Err() == nil {
			numGood++
		}
		return false
	}

	var srch runnable.Search
	traces := runnable.NewTraceWatcher()
	srch.Run(gg.SearchInit, resultFunc, traces)

	if numGood == 0 {
		t.Logf("didn't find any solution")
		t.Fail()
	} else if numGood > 1 {
		t.Logf("found too many solutions: %v", numGood)
		t.Fail()
	}

	if t.Failed() {
		gg = runnable.NewStepGen(word.NewPlanProblem(&prob, true))
		planf(gg.PlanProblem, gg.LoggedGen(), true)
		srch.Run(gg.SearchInit, resultFunc, runnable.Watchers([]runnable.SearchWatcher{
			traces,
			runnable.DebugWatcher{
				Logf: logf,
			},
		}))
	}
}

func benchGogenPlan(b *testing.B, planf word.PlanFunc, w1, w2, w3 string) {
	var prob word.Problem
	if err := prob.Setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}
	for n := 0; n < b.N; n++ {
		gg := runnable.NewStepGen(word.NewPlanProblem(&prob, false))
		planf(gg.PlanProblem, gg, false)
	}
}

func benchGogenRun(b *testing.B, planf word.PlanFunc, w1, w2, w3 string) {
	var prob word.Problem
	if err := prob.Setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	gg := runnable.NewStepGen(word.NewPlanProblem(&prob, false))
	planf(gg.PlanProblem, gg, false)

	for n := 0; n < b.N; n++ {
		var srch runnable.Search
		numGood := 0
		srch.Run(
			gg.SearchInit,
			func(sol *runnable.Solution) bool {
				if sol.Err() == nil {
					numGood++
				}
				return false
			},
			nil)
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
