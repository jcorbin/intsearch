package stackvm

import "fmt"

var (
	// Each starts an `Each PRED... Do BODY... End`: repeatedly run PRED and
	// BODY as long as PRED is true.
	Each = _each{"Each", condJz}

	// Much starts a `Much PRED... Then NEXT... [Else EXIT] End`: if PRED is
	// true, fork NEXT and loop; otherwise run any EXIT.
	Much = _much{"Much", condJz, condFnz}

	// Do starts the body of a Each.
	Do = _do{}
)

type _each struct {
	name  string
	econd func(off int) Op
}
type _much struct {
	name  string
	econd func(off int) Op
	fcond func(off int) Op
}
type _do struct{}

func (e _each) compile() consumer { return &_eachCtx{_each: e, cur: []Op{}} }
func (m _much) compile() consumer { return &_muchCtx{_much: m, cur: []Op{}} }

type _eachCtx struct {
	_each
	cur, pred, body []Op
}

type _muchCtx struct {
	_much
	cur, pred, next, exit []Op
}

func (ec *_eachCtx) consume(x interface{}) ([]Op, error) {
	switch x.(type) {
	case _do:
		ec.pred, ec.cur = ec.cur, []Op{}
	case _end:
		ec.body, ec.cur = ec.cur, nil
		return ec.finalize(), nil
	}
	return nil, nil
}

func (mc *_muchCtx) consume(x interface{}) ([]Op, error) {
	switch x.(type) {
	case _then:
		mc.pred, mc.cur = mc.cur, []Op{}
	case _else:
		mc.next, mc.cur = mc.cur, []Op{}
	case _end:
		if mc.next == nil {
			mc.next = mc.cur
		} else {
			mc.exit = mc.cur
		}
		mc.cur = nil
		return mc.finalize(), nil
	}
	return nil, nil
}

func (ec *_eachCtx) consumeOps(ops ...Op) error {
	if ec.cur == nil {
		return fmt.Errorf("unexpected %v ops after end of %s", ops, ec.name)
	}
	ec.cur = append(ec.cur, ops...)
	return nil
}

func (mc *_muchCtx) consumeOps(ops ...Op) error {
	if mc.cur == nil {
		return fmt.Errorf("unexpected %v ops after end of %s", ops, mc.name)
	}
	mc.cur = append(mc.cur, ops...)
	return nil
}

func (ec *_eachCtx) finalize() []Op {
	nPred := len(ec.pred) + 1
	nBody := len(ec.body) + 1
	ops := make([]Op, 0, nPred+nBody)
	ops = append(ops, ec.pred...)
	ops = append(ops, ec.econd(nBody))
	ops = append(ops, ec.body...)
	ops = append(ops, Jmp(-nPred-nBody))
	return ops
}

func (mc *_muchCtx) finalize() []Op {
	nExit := len(mc.exit)
	if nExit > 0 {
		panic("not implemented")
	}

	nNext := len(mc.next)
	nPred := len(mc.pred)
	nHead := nPred + nNext
	ops := make([]Op, 0, 1+nHead+5)
	ops = append(ops, Jmp(nNext))
	ops = append(ops, mc.next...)
	ops = append(ops, mc.pred...)
	ops = append(ops,
		Dup, mc.econd(2),
		mc.fcond(-3-nHead), Jmp(1),
		Pop,
	)

	return ops
}

func (ec *_eachCtx) String() string {
	var next string
	s := ec.name
	if len(ec.pred) > 0 {
		s += fmt.Sprintf(" %v", ec.pred)
		next = "Do"
	}
	if len(ec.body) > 0 {
		s += fmt.Sprintf(" Do %v", ec.body)
		next = ""
	}

	if ec.cur != nil {
		if next == "" {
			s += fmt.Sprintf(" %v...", ec.cur)
		} else {
			s += fmt.Sprintf(" %s %v...", next, ec.cur)
		}
	} else {
		s += " End"
	}
	return s
}

func (e _each) String() string { return e.name }
func (m _much) String() string { return m.name }
func (d _do) String() string   { return "Do" }

// TODO: can't do break or continue currently, since Compile delegates
// directly rather than intermediating
// Do ... Break ... Continue ... End
