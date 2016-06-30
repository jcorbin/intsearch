package runnable_test

import (
	"testing"

	"github.com/jcorbin/intsearch/internal/gen_testing"
	"github.com/jcorbin/intsearch/runnable"
	"github.com/jcorbin/intsearch/word"
)

func TestStepGen_prunedBrute(t *testing.T) {
	gen_testing.RunGenTest(t, runnable.NewStepGen, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkStepGenPlan_prunedBrute(b *testing.B) {
	gen_testing.BenchGenPlan(b, runnable.NewStepGen, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkStepGenRun_prunedBrute(b *testing.B) {
	gen_testing.BenchGenRun(b, runnable.NewStepGen, word.PlanPrunedBrute, "send", "more", "money")
}

func TestStepGen_bottomUp(t *testing.T) {
	gen_testing.RunGenTest(t, runnable.NewStepGen, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkStepGenPlan_bottomUp(b *testing.B) {
	gen_testing.BenchGenPlan(b, runnable.NewStepGen, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkStepGenRun_bottomUp(b *testing.B) {
	gen_testing.BenchGenRun(b, runnable.NewStepGen, word.PlanBottomUp, "send", "more", "money")
}

func TestStepGen_topDown(t *testing.T) {
	gen_testing.RunGenTest(t, runnable.NewStepGen, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkStepGenPlan_topDown(b *testing.B) {
	gen_testing.BenchGenPlan(b, runnable.NewStepGen, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkStepGenRun_topDown(b *testing.B) {
	gen_testing.BenchGenRun(b, runnable.NewStepGen, word.PlanTopDown, "send", "more", "money")
}
