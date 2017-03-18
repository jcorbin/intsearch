package main

import (
	"fmt"
	"reflect"
	"runtime"
)

var (
	alloc = _alloc{}
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

type _alloc struct{}
type _load struct{}
type _store struct{}
type _dup struct{}
type _swap struct{}

type push byte

type label string

type labelRef struct {
	name label
	ref  func(int) machStep
}

func (op label) comeFrom(ref func(int) machStep) interface{} {
	return labelRef{op, ref}
}

func (op label) annotate(ann map[string]int, i int) {
	ann[string(op)] = i
}

func (op labelRef) resolve(ann map[string]int, i int) machStep {
	n := string(op.name)
	if j, def := ann[n]; def {
		return op.ref(j - i - 1)
	}
	return nil
}

type _halt struct{ err error }
type _hnz struct{ err error }
type _hz struct{ err error }

func halt(err error) interface{} { return _halt{err} }
func hnz(err error) interface{}  { return _hnz{err} }
func hz(err error) interface{}   { return _hz{err} }

type fnz int
type jmp int
type jnz int
type jz int

func fnzFrom(off int) machStep { return fnz(off) }
func jmpFrom(off int) machStep { return jmp(off) }
func jzFrom(off int) machStep  { return jz(off) }
func jnzFrom(off int) machStep { return jnz(off) }

type _add struct{}
type _sub struct{}
type _mod struct{}
type _div struct{}

type _lt struct{}
type _eq struct{}

// Op Stringers

func (op _alloc) String() string   { return "alloc" }
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
func (op push) String() string     { return fmt.Sprintf("push %d", byte(op)) }
func (op fnz) String() string      { return fmt.Sprintf("fnz %d", int(op)) }
func (op jmp) String() string      { return fmt.Sprintf("jmp %d", int(op)) }
func (op jnz) String() string      { return fmt.Sprintf("jnz %d", int(op)) }
func (op jz) String() string       { return fmt.Sprintf("jz %d", int(op)) }
func (op labelRef) String() string { return fmt.Sprintf("%v <- %v", funcName(op.ref), op.name) }
func (op _halt) String() string    { return fmt.Sprintf("halt %v", op.err) }
func (op _hnz) String() string     { return fmt.Sprintf("hnz %v", op.err) }
func (op _hz) String() string      { return fmt.Sprintf("hz %v", op.err) }

func funcName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
