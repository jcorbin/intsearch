package stackvm

import "fmt"

func condJnz(off int) Op { return Jnz(off) }
func condJz(off int) Op  { return Jz(off) }

var (
	// If starts an `If PRED... Then BODY... [Else BODY...] End`.
	If = _guard{"If", condJz}

	// Unless starts an `Unless PRED... Then BODY... [Else BODY...] End`.
	Unless = _guard{"Unless", condJnz}

	// Then starts the body of an If or Unless.
	Then = _then{}

	// Else starts the alternate body of an If or Unless
	Else = _else{}

	// End ends a piece of control structure.
	End = _end{}
)

type _guard struct {
	name  string
	acond func(off int) Op
}
type _then struct{}
type _else struct{}
type _end struct{}

func (g _guard) compile() consumer { return &_guardCtx{_guard: g, cur: []Op{}} }
func (g _guard) String() string    { return g.name }

type _guardCtx struct {
	_guard
	cur, pred, body, alt []Op
}

func (gc *_guardCtx) consume(x interface{}) ([]Op, error) {
	switch x.(type) {
	case _then:
		gc.pred, gc.cur = gc.cur, []Op{}
	case _else:
		gc.body, gc.cur = gc.cur, []Op{}
	case _end:
		if gc.body == nil {
			gc.body = gc.cur
		} else {
			gc.alt = gc.cur
		}
		gc.cur = nil
		return gc.finalize(), nil
	}
	return nil, nil
}

func (gc *_guardCtx) consumeOps(ops ...Op) error {
	if gc.cur == nil {
		return fmt.Errorf("unexpected %v ops after end of %s", ops, gc.name)
	}
	gc.cur = append(gc.cur, ops...)
	return nil
}

func (gc *_guardCtx) finalize() []Op {
	if len(gc.alt) == 0 {
		ops := make([]Op, 0, len(gc.pred)+1+len(gc.body))
		ops = append(ops, gc.pred...)
		ops = append(ops, gc.acond(len(gc.body)))
		ops = append(ops, gc.body...)
		return ops
	}
	ops := make([]Op, 0, len(gc.pred)+1+len(gc.body)+1+len(gc.alt))
	ops = append(ops, gc.pred...)
	ops = append(ops, gc.acond(len(gc.body)+1))
	ops = append(ops, gc.body...)
	ops = append(ops, Jmp(len(gc.alt)))
	ops = append(ops, gc.alt...)
	return ops
}

func (gc *_guardCtx) String() string {
	var next string
	s := gc.name
	if len(gc.pred) > 0 {
		s += fmt.Sprintf(" %v", gc.pred)
		next = "Then"
	}
	if len(gc.body) > 0 {
		s += fmt.Sprintf(" Then %v", gc.body)
		next = "Else"
	}
	if len(gc.alt) > 0 {
		s += fmt.Sprintf(" Else %v", gc.alt)
		next = ""
	}

	if gc.cur != nil {
		if next == "" {
			s += fmt.Sprintf(" %v...", gc.cur)
		} else {
			s += fmt.Sprintf(" %s %v...", next, gc.cur)
		}
	} else {
		s += " End"
	}
	return s
}
func (t _then) String() string { return "Then" }
func (e _else) String() string { return "Else" }
func (d _end) String() string  { return "End" }

// TODO switch, else/if chain
