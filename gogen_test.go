package main

import "testing"

func TestGogenSendMoreMoney(t *testing.T) {
	var (
		prob problem
		gg   = goGen{
			verified: true,
		}
		traces = newTraceWatcher()
		srch   search
	)

	if err := prob.setup("send", "more", "money"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	planBottomUp(&prob, &gg)

	numGood := 0

	initFunc := func(emit emitFunc) {
		emit(newSolution(&prob, gg.steps, emit))
	}

	resultFunc := func(sol *solution) {
		if sol.err == errVerifyFailed {
			t.Logf("!!! invalid solution found: %v %v", sol, sol.letterMapping())
			for i, soli := range sol.trace {
				t.Logf("trace[%v]: %v %s", i, soli, soli.letterMapping())
			}
			t.Fail()
		} else if sol.err != nil {
			// normal dead end result, discard
			return
		} else {
			numGood++
		}
	}

	srch.run(100000, initFunc, resultFunc, traces)

	if numGood == 0 {
		t.Logf("didn't find any solution")
		t.Fail()
	} else if numGood > 1 {
		t.Logf("found too many solutions: %v", numGood)
		t.Fail()
	}
}

func BenchmarkPlan(b *testing.B) {
	var prob problem
	if err := prob.setup("send", "more", "money"); err != nil {
		b.Fatalf("seutp failed: %v", err)
	}
	for n := 0; n < b.N; n++ {
		var gg goGen
		planBottomUp(&prob, &gg)
	}
}

func BenchmarkRun(b *testing.B) {
	var (
		prob problem
		gg   goGen
	)
	if err := prob.setup("send", "more", "money"); err != nil {
		b.Fatalf("seutp failed: %v", err)
	}
	planBottomUp(&prob, &gg)

	for n := 0; n < b.N; n++ {
		var srch search
		numGood := 0
		srch.run(
			100000,
			func(emit emitFunc) {
				emit(newSolution(&prob, gg.steps, emit))
			},
			func(sol *solution) {
				if sol.err == nil {
					numGood++
				} else {
					// normal dead end result, discard
					return
				}
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
