package word

// MultiGen implements a simple one-to-many SolutionGen.  Its primary use case
// is to combine an observability gen with a primary concrete gen.
type MultiGen []SolutionGen

// Problem returns the first non-nil problem from a gen.
func (mg MultiGen) Problem() *PlanProblem {
	for _, gen := range mg {
		if prob := gen.Problem(); prob != nil {
			return prob
		}
	}
	return nil
}

// Logf passes the log message to all gens.
func (mg MultiGen) Logf(format string, args ...interface{}) error {
	for _, gen := range mg {
		gen.Logf(format, args...)
	}
	return nil // TODO: multiError fwiw
}

// Init calls each gen's Init method.
func (mg MultiGen) Init(desc string) {
	for _, gen := range mg {
		gen.Init(desc)
	}
}

// Fork forks each gen, and returns a new MultiGen containing all forked gens.
func (mg MultiGen) Fork(prob *PlanProblem, name, alt, cont string) SolutionGen {
	altGen := make([]SolutionGen, len(mg))
	for i, gen := range mg {
		altGen[i] = gen.Fork(prob, name, alt, cont)
	}
	return MultiGen(altGen)
}

// Fix calls each gen's Fix method.
func (mg MultiGen) Fix(c byte, v int) {
	for _, gen := range mg {
		gen.Fix(c, v)
	}
}

// ComputeSum calls each gen's ComputeSum method.
func (mg MultiGen) ComputeSum(col *Column) {
	for _, gen := range mg {
		gen.ComputeSum(col)
	}
}

// ComputeFirstSummand calls each gen's ComputeFirstSummand method.
func (mg MultiGen) ComputeFirstSummand(col *Column) {
	for _, gen := range mg {
		gen.ComputeFirstSummand(col)
	}
}

// ComputeSecondSummand calls each gen's ComputeSecondSummand method.
func (mg MultiGen) ComputeSecondSummand(col *Column) {
	for _, gen := range mg {
		gen.ComputeSecondSummand(col)
	}
}

// ChooseRange calls each gen's ChooseRange method.
func (mg MultiGen) ChooseRange(c byte, min, max int) {
	for _, gen := range mg {
		gen.ChooseRange(c, min, max)
	}
}

// CheckColumn calls each gen's CheckColumn method.
func (mg MultiGen) CheckColumn(col *Column, err error) {
	for _, gen := range mg {
		gen.CheckColumn(col, err)
	}
}

// Verify calls each gen's Verify method.
func (mg MultiGen) Verify() {
	for _, gen := range mg {
		gen.Verify()
	}
}

// Check calls each gen's Check method.
func (mg MultiGen) Check(err error) {
	for _, gen := range mg {
		gen.Check(err)
	}
}

// Finish calls each gen's Finish method.
func (mg MultiGen) Finish() {
	for _, gen := range mg {
		gen.Finish()
	}
}

// Finalize calls each gen's Finalize method, and returns the only non-nil
// plan; panics if more than one gen.Finalize() returns a non-nil plan.
func (mg MultiGen) Finalize() Plan {
	var plan Plan
	for _, gen := range mg {
		p := gen.Finalize()
		if p != nil {
			if plan != nil {
				panic("more than one concrete plan")
			}
			plan = p
		}
	}
	return plan
}
