package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestGogenSendMoreMoney(t *testing.T) {
	var prob problem
	if err := prob.setup("send", "more", "money"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	decs := func(args ...interface{}) string {
		dec := gg.decorate(args)
		if len(dec) == 0 {
			return ""
		}
		return fmt.Sprintf("  // %s", strings.Join(dec, ", "))
	}

	gg := newGoGen(newPlanProblem(&prob))
	gg.verified = true
	gg.planProblem.plan(gg)

	numGood := 0

	resultFunc := func(sol *solution) bool {
		if sol.err == errVerifyFailed {
			t.Logf("!!! invalid solution found: %v %s%s", sol, sol.letterMapping(), decs(sol))
			for i, soli := range sol.trace {
				t.Logf("trace[%v]: %v %s%s", i, soli, soli.letterMapping(), decs(soli))
			}
			t.Fail()
		} else if sol.err == nil {
			numGood++
		}
		return false
	}

	var srch search
	traces := newTraceWatcher()
	srch.run(100000, gg.searchInit, resultFunc, traces)

	if numGood == 0 {
		t.Logf("didn't find any solution")
		t.Fail()
	} else if numGood > 1 {
		t.Logf("found too many solutions: %v", numGood)
		t.Fail()
	}

	if t.Failed() {
		gg := newGoGen(newPlanProblem(&prob))
		gg.verified = true
		gg.planProblem.plan(gg.loggedGen())
		srch.run(100000, gg.searchInit, resultFunc, watchers([]searchWatcher{
			traces,
			debugWatcher{},
		}))
	}
}

func BenchmarkPlan(b *testing.B) {
	var prob problem
	if err := prob.setup("send", "more", "money"); err != nil {
		b.Fatalf("seutp failed: %v", err)
	}
	for n := 0; n < b.N; n++ {
		gg := newGoGen(newPlanProblem(&prob))
		gg.planProblem.plan(gg)
	}
}

func BenchmarkRun(b *testing.B) {
	var prob problem
	if err := prob.setup("send", "more", "money"); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	gg := newGoGen(newPlanProblem(&prob))
	gg.planProblem.plan(gg)

	for n := 0; n < b.N; n++ {
		var srch search
		numGood := 0
		srch.run(
			100000,
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
