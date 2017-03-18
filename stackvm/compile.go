package stackvm

import "errors"

var errUnexpectedEnd = errors.New("unexpected end of input")

// Compile compiles code into a new stack machine.
func Compile(codes []interface{}) (*Mach, error) {
	var ctx = []consumer{&prog{}}

	for _, code := range codes {
		switch val := code.(type) {

		case compiler:
			ctx = append(ctx, val.compile())
			continue

		case consumer:
			ctx = append(ctx, val)
			continue

		case Op:
			if err := ctx[len(ctx)-1].consumeOps(val); err != nil {
				return nil, err
			}

		default:
			ops, err := ctx[len(ctx)-1].consume(code)
			if err != nil {
				return nil, err
			}
			if len(ops) > 0 {
				ctx = ctx[:len(ctx)-1]
				if err := ctx[len(ctx)-1].consumeOps(ops...); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(ctx) > 1 {
		return nil, errUnexpectedEnd
	}

	prog, err := ctx[0].consume(nil)
	if err != nil {
		return nil, err
	}

	return &Mach{
		prog: prog,
		heap: _size,
	}, nil
}

type compiler interface {
	compile() consumer
}

type consumer interface {
	consume(interface{}) ([]Op, error)
	consumeOps(...Op) error
}

type prog struct{ ops []Op }

func (p *prog) consume(code interface{}) ([]Op, error) {
	if code == nil {
		ops := p.ops
		p.ops = nil
		return ops, nil
	}
	if op, ok := code.(Op); ok {
		p.ops = append(p.ops, op)
	}
	return nil, nil
}

func (p *prog) consumeOps(ops ...Op) error {
	p.ops = append(p.ops, ops...)
	return nil
}
