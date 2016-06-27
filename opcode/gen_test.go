package opcode_test

import (
	"testing"

	"github.com/jcorbin/intsearch/internal/gen_testing"
	"github.com/jcorbin/intsearch/opcode"
	"github.com/jcorbin/intsearch/word"
)

func TestCodeGen_prunedBrute(t *testing.T) {
	gen_testing.RunGenTest(t, opcode.NewCodeGen, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkCodeGenPlan_prunedBrute(b *testing.B) {
	gen_testing.BenchGenPlan(b, opcode.NewCodeGen, word.PlanPrunedBrute, "send", "more", "money")
}

func BenchmarkCodeGenRun_prunedBrute(b *testing.B) {
	gen_testing.BenchGenRun(b, opcode.NewCodeGen, word.PlanPrunedBrute, "send", "more", "money")
}

func TestCodeGen_bottomUp(t *testing.T) {
	gen_testing.RunGenTest(t, opcode.NewCodeGen, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkCodeGenPlan_bottomUp(b *testing.B) {
	gen_testing.BenchGenPlan(b, opcode.NewCodeGen, word.PlanBottomUp, "send", "more", "money")
}

func BenchmarkCodeGenRun_bottomUp(b *testing.B) {
	gen_testing.BenchGenRun(b, opcode.NewCodeGen, word.PlanBottomUp, "send", "more", "money")
}

func TestCodeGen_topDown(t *testing.T) {
	gen_testing.RunGenTest(t, opcode.NewCodeGen, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkCodeGenPlan_topDown(b *testing.B) {
	gen_testing.BenchGenPlan(b, opcode.NewCodeGen, word.PlanTopDown, "send", "more", "money")
}

func BenchmarkCodeGenRun_topDown(b *testing.B) {
	gen_testing.BenchGenRun(b, opcode.NewCodeGen, word.PlanTopDown, "send", "more", "money")
}
