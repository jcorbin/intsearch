package main

var (
	load  = _load{}
	store = _store{}
	dup   = _dup{}
	swap  = _swap{}

	add = _add{}
	sub = _sub{}
	mod = _mod{}
	div = _div{}

	lt = _lt{}
	eq = _eq{}
)

type _load struct{}
type _store struct{}
type _dup struct{}
type _swap struct{}

type push int

type label string

type labelRef struct {
	name label
	ref  func(int) interface{}
}

func (name label) comeFrom(ref func(int) interface{}) interface{} {
	return labelRef{name, ref}
}

func (ref labelRef) compile(steps []interface{}) []interface{} {
	labels := make(map[label]int, len(steps))
	pending := make(map[label][]int, len(steps))
	for i, step := range steps {
		switch s := step.(type) {
		case label:
			labels[s] = i
			if pend := pending[s]; len(pend) > 0 {
				for _, refi := range pend {
					steps[refi] = steps[refi].(labelRef).ref(i)
				}
				delete(pending, s)
			}
		case labelRef:
			if j, def := labels[s.name]; def {
				steps[i] = s.ref(j)
				continue
			}
			pending[s.name] = append(pending[s.name], i)
		}
	}
	return steps
}

type _halt struct{ error }
type _hnz struct{ error }
type _hz struct{ error }

func halt(err error) interface{} { return _halt{err} }
func hnz(err error) interface{}  { return _hnz{err} }
func hz(err error) interface{}   { return _hz{err} }

type fnz int
type jnz int
type jz int

func fnzFrom(off int) interface{} { return fnz(off) }
func jzFrom(off int) interface{}  { return jz(off) }
func jnzFrom(off int) interface{} { return jnz(off) }

type _add struct{}
type _sub struct{}
type _mod struct{}
type _div struct{}

type _lt struct{}
type _eq struct{}
