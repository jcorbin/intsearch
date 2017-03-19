package stackvm

import "fmt"

var (
	// While starts a `While PRED... Loop BODY... End`.
	While = _each{"While", condJz}

	// Until starts an `Until PRED... Loop BODY... End`.
	Until = _each{"Until", condJnz}

	// Loop starts the body of a While or Until.
	Loop = _loop{}
)

type _each struct {
	name  string
	econd func(off int) Op
}
type _loop struct{}

func (e _each) compile() consumer { return &_eachCtx{_each: e, cur: []Op{}} }

type _eachCtx struct {
	_each
	cur, pred, body []Op
}

func (ec *_eachCtx) consume(x interface{}) ([]Op, error) {
	switch x.(type) {
	case _loop:
		ec.pred, ec.cur = ec.cur, []Op{}
	case _end:
		ec.body, ec.cur = ec.cur, nil
		return ec.finalize(), nil
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

func (ec *_eachCtx) finalize() []Op {
	n := len(ec.pred) + 1 + len(ec.body) + 1
	ops := make([]Op, 0, n)
	ops = append(ops, ec.pred...)
	ops = append(ops, ec.econd(len(ec.body)+1))
	ops = append(ops, ec.body...)
	ops = append(ops, Jmp(-n))
	return ops
}

func (ec *_eachCtx) String() string {
	var next string
	s := ec.name
	if len(ec.pred) > 0 {
		s += fmt.Sprintf(" %v", ec.pred)
		next = "Loop"
	}
	if len(ec.body) > 0 {
		s += fmt.Sprintf(" Loop %v", ec.body)
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
func (l _loop) String() string { return "Loop" }

// TODO: can't do break or continue currently, since Compile delegates
// directly rather than intermediating
// Loop ... Break ... Continue ... End
