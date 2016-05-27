package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestGogen_prunedBrute(t *testing.T) {
	runGogenTest(t, planPrunedBrute, "send", "more", "money")
}

func BenchmarkGogenPlan_prunedBrute(b *testing.B) {
	benchGogenPlan(b, planPrunedBrute, "send", "more", "money")
}

func BenchmarkGogenRun_prunedBrute(b *testing.B) {
	benchGogenRun(b, planPrunedBrute, "send", "more", "money")
}

func TestGogen_bottomUp(t *testing.T) {
	runGogenTest(t, planBottomUp, "send", "more", "money")
}

func BenchmarkGogenPlan_bottomUp(b *testing.B) {
	benchGogenPlan(b, planBottomUp, "send", "more", "money")
}

func BenchmarkGogenRun_bottomUp(b *testing.B) {
	benchGogenRun(b, planBottomUp, "send", "more", "money")
}

func TestGogen_topDown(t *testing.T) {
	runGogenTest(t, planTopDown, "send", "more", "money")
}

func BenchmarkGogenPlan_topDown(b *testing.B) {
	benchGogenPlan(b, planTopDown, "send", "more", "money")
}

func BenchmarkGogenRun_topDown(b *testing.B) {
	benchGogenRun(b, planTopDown, "send", "more", "money")
}

func runGogenTest(t *testing.T, planf planFunc, w1, w2, w3 string) {
	var prob problem
	if err := prob.setup(w1, w2, w3); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	var gg *goGen

	logf := func(format string, args ...interface{}) {
		dec := gg.decorate(args)
		if len(dec) > 0 {
			format = fmt.Sprintf("%s  // %s", format, strings.Join(dec, ", "))
		}
		t.Logf(format, args...)
	}

	gg = newGoGen(newPlanProblem(&prob, false))
	planf(gg.planProblem, gg, true)

	numGood := 0

	resultFunc := func(sol *solution) bool {
		if isVerifyError(sol.err) {
			logf("!!! invalid solution found: %v %s", sol, sol.letterMapping())
			for i, soli := range sol.trace {
				logf("trace[%v]: %v %s", i, soli, soli.letterMapping())
			}
			t.Fail()
		} else if sol.err == nil {
			numGood++
		}
		return false
	}

	var srch search
	traces := newTraceWatcher()
	srch.run(gg.searchInit, resultFunc, traces)

	if numGood == 0 {
		t.Logf("didn't find any solution")
		t.Fail()
	} else if numGood > 1 {
		t.Logf("found too many solutions: %v", numGood)
		t.Fail()
	}

	if t.Failed() {
		gg = newGoGen(newPlanProblem(&prob, true))
		planf(gg.planProblem, gg.loggedGen(), true)
		srch.run(gg.searchInit, resultFunc, watchers([]searchWatcher{
			traces,
			debugWatcher{
				logf: logf,
			},
		}))
	}
}

func benchGogenPlan(b *testing.B, planf planFunc, w1, w2, w3 string) {
	var prob problem
	if err := prob.setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}
	for n := 0; n < b.N; n++ {
		gg := newGoGen(newPlanProblem(&prob, false))
		planf(gg.planProblem, gg, false)
		gg.compile()
	}
}

func benchGogenRun(b *testing.B, planf planFunc, w1, w2, w3 string) {
	var prob problem
	if err := prob.setup(w1, w2, w3); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	gg := newGoGen(newPlanProblem(&prob, false))
	planf(gg.planProblem, gg, false)
	gg.compile()

	for n := 0; n < b.N; n++ {
		var srch search
		numGood := 0
		srch.run(
			gg.searchInit,
			func(sol *solution) bool {
				if sol.err == nil {
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
