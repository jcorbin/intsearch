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

	var gg *goGen

	logf := func(format string, args ...interface{}) {
		dec := gg.decorate(args)
		if len(dec) > 0 {
			format = fmt.Sprintf("%s  // %s", format, strings.Join(dec, ", "))
		}
		t.Logf(format, args...)
	}

	gg = newGoGen(newPlanProblem(&prob), true)
	gg.planProblem.plan(gg)

	numGood := 0

	resultFunc := func(sol *solution) bool {
		if sol.err == errVerifyFailed {
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
	srch.run(100000, gg.searchInit, resultFunc, traces)

	if numGood == 0 {
		t.Logf("didn't find any solution")
		t.Fail()
	} else if numGood > 1 {
		t.Logf("found too many solutions: %v", numGood)
		t.Fail()
	}

	if t.Failed() {
		gg = newGoGen(newPlanProblem(&prob), true)
		gg.planProblem.plan(gg.loggedGen())
		srch.run(100000, gg.searchInit, resultFunc, watchers([]searchWatcher{
			traces,
			debugWatcher{
				logf: logf,
			},
		}))
	}
}

func BenchmarkPlan(b *testing.B) {
	var prob problem
	if err := prob.setup("send", "more", "money"); err != nil {
		b.Fatalf("setup failed: %v", err)
	}
	for n := 0; n < b.N; n++ {
		gg := newGoGen(newPlanProblem(&prob), false)
		gg.planProblem.plan(gg)
		gg.compile()
	}
}

func BenchmarkRun(b *testing.B) {
	var prob problem
	if err := prob.setup("send", "more", "money"); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	gg := newGoGen(newPlanProblem(&prob), false)
	gg.planProblem.plan(gg)
	gg.compile()

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
