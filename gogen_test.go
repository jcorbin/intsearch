package main

import "testing"

func TestGogenSendMoreMoney(t *testing.T) {
	var (
		prob problem
		gg   = goGen{
			verified: true,
		}
		traces = newTraceWatcher()
		srch   = search{
			watcher: traces,
		}
	)

	if err := prob.plan("send", "more", "money", &gg); err != nil {
		t.Fatalf("plan failed: %v", err)
	}

	numGood := 0

	srch.run(
		100000,
		func(emit emitFunc) {
			emit(newSolution(&prob, gg.steps, emit))
		},
		func(sol *solution) {
			if sol.err == errVerifyFailed {
				t.Logf("!!! invalid solution found: %v %v", sol, sol.letterMapping())
				for i, soli := range traces[sol] {
					t.Logf("trace[%v]: %v %s", i, soli, soli.letterMapping())
				}
				t.Fail()
			} else if sol.err != nil {
				// normal dead end result, discard
				return
			} else {
				numGood++
			}
		})

	if numGood == 0 {
		t.Fatalf("didn't find any solution")
	} else if numGood > 1 {
		t.Fatalf("found too many solutions: %v", numGood)
	}
}
