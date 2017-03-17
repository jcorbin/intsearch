package main

import "fmt"

type step interface {
}

var (
	alloc = _alloc{}
	load  = _load{}
	store = _store{}
	dup   = _dup{}
	swap  = _swap{}
	add   = _add{}
	sub   = _sub{}
	div   = _div{}
	mod   = _mod{}
	lt    = _lt{}
	lte   = _lte{}
	eq    = _eq{}
	neq   = _neq{}
	gt    = _gt{}
	gte   = _gte{}
)

type rem struct{ s string }
type com struct {
	step
	s string
}
type halt struct{ error }
type push int
type jmp int
type jnz int
type jz int
type fork int
type fnz int
type fz int
type _alloc struct{}
type _load struct{}
type _store struct{}
type _dup struct{}
type _swap struct{}
type _add struct{}
type _sub struct{}
type _div struct{}
type _mod struct{}
type _lt struct{}
type _lte struct{}
type _eq struct{}
type _neq struct{}
type _gt struct{}
type _gte struct{}

func (r rem) String() string     { return fmt.Sprintf("-- %s", r.s) }
func (c com) String() string     { return fmt.Sprintf("%v -- %s", c.step, c.s) }
func (op halt) String() string   { return fmt.Sprintf("halt %v", op.error) }
func (op push) String() string   { return fmt.Sprintf("push %d", int(op)) }
func (op jmp) String() string    { return fmt.Sprintf("jmp %d", int(op)) }
func (op jnz) String() string    { return fmt.Sprintf("jnz %d", int(op)) }
func (op jz) String() string     { return fmt.Sprintf("jz %d", int(op)) }
func (op fork) String() string   { return fmt.Sprintf("fork %d", int(op)) }
func (op fnz) String() string    { return fmt.Sprintf("fnz %d", int(op)) }
func (op fz) String() string     { return fmt.Sprintf("fz %d", int(op)) }
func (op _alloc) String() string { return "alloc" }
func (op _load) String() string  { return "load" }
func (op _store) String() string { return "store" }
func (op _dup) String() string   { return "dup" }
func (op _swap) String() string  { return "swap" }
func (op _add) String() string   { return "add" }
func (op _sub) String() string   { return "sub" }
func (op _div) String() string   { return "div" }
func (op _mod) String() string   { return "mod" }
func (op _lt) String() string    { return "lt" }
func (op _lte) String() string   { return "lte" }
func (op _eq) String() string    { return "eq" }
func (op _neq) String() string   { return "neq" }
func (op _gt) String() string    { return "gt" }
func (op _gte) String() string   { return "gte" }

func remf(pat string, args ...interface{}) rem {
	return rem{fmt.Sprintf(pat, args...)}
}

func comd(s step, pat string, args ...interface{}) com {
	return com{
		step: s,
		s:    fmt.Sprintf(pat, args...),
	}
}
