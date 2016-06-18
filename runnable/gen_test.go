package runnable_test

import (
	"testing"

	"github.com/jcorbin/intsearch/internal/gen_testing"
	"github.com/jcorbin/intsearch/runnable"
	"github.com/jcorbin/intsearch/word"
)

func stepGenF(prob *word.PlanProblem) word.SolutionGen {
	return runnable.NewStepGen(prob)
}

func TestStepGen_prunedBrute(t *testing.T) {
	gen_testing.RunGenTest(t, stepGenF, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkStepGenPlan_prunedBrute(b *testing.B) {
	gen_testing.BenchGenPlan(b, stepGenF, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkStepGenRun_prunedBrute(b *testing.B) {
	gen_testing.BenchGenRun(b, stepGenF, word.PlanPrunedBrute, "send", "more", "money")
}

func TestStepGen_bottomUp(t *testing.T) {
	gen_testing.RunGenTest(t, stepGenF, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkStepGenPlan_bottomUp(b *testing.B) {
	gen_testing.BenchGenPlan(b, stepGenF, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkStepGenRun_bottomUp(b *testing.B) {
	gen_testing.BenchGenRun(b, stepGenF, word.PlanBottomUp, "send", "more", "money")
}

func TestStepGen_topDown(t *testing.T) {
	gen_testing.RunGenTest(t, stepGenF, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkStepGenPlan_topDown(b *testing.B) {
	gen_testing.BenchGenPlan(b, stepGenF, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkStepGenRun_topDown(b *testing.B) {
	gen_testing.BenchGenRun(b, stepGenF, word.PlanTopDown, "send", "more", "money")
}
