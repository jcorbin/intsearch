package main

import (
	"fmt"
	"reflect"
	"runtime"
)

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

func (op label) comeFrom(ref func(int) interface{}) interface{} {
	return labelRef{op, ref}
}

func compileLabelRefs(steps []interface{}) []interface{} {
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

// Op Stringers

func (op _load) String() string    { return "load" }
func (op _store) String() string   { return "store" }
func (op _dup) String() string     { return "dup" }
func (op _swap) String() string    { return "swap" }
func (op _add) String() string     { return "add" }
func (op _sub) String() string     { return "sub" }
func (op _mod) String() string     { return "mod" }
func (op _div) String() string     { return "div" }
func (op _lt) String() string      { return "lt" }
func (op _eq) String() string      { return "eq" }
func (op label) String() string    { return ":" + string(op) }
func (op push) String() string     { return fmt.Sprintf("push %d", int(op)) }
func (op fnz) String() string      { return fmt.Sprintf("fnz %d", int(op)) }
func (op jnz) String() string      { return fmt.Sprintf("jnz %d", int(op)) }
func (op jz) String() string       { return fmt.Sprintf("jz %d", int(op)) }
func (op labelRef) String() string { return fmt.Sprintf("%v <- %v", op.name, funcName(op.ref)) }
func (op _halt) String() string    { return fmt.Sprintf("halt %v", op.error) }
func (op _hnz) String() string     { return fmt.Sprintf("hnz %v", op.error) }
func (op _hz) String() string      { return fmt.Sprintf("hz %v", op.error) }

func funcName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
