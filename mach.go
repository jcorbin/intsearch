package main

import "errors"

type machStep interface {
	step(*mach) error
}

type mach struct {
	prog []machStep
	pc   int
}

func newMach(prog []machStep) *mach {
	return &mach{
		prog: prog,
	}
}

func (m *mach) run() error {
	for m.pc < len(m.prog) {
		ms := m.prog[m.pc]
		if err := ms.step(m); err != nil {
			return err
		}
		m.pc++
	}
	return nil
}

var errUnimplemented = errors.New("op unimplemented")

func (op _load) step(_ *mach) error    { return errUnimplemented }
func (op _store) step(_ *mach) error   { return errUnimplemented }
func (op _dup) step(_ *mach) error     { return errUnimplemented }
func (op _swap) step(_ *mach) error    { return errUnimplemented }
func (op _add) step(_ *mach) error     { return errUnimplemented }
func (op _sub) step(_ *mach) error     { return errUnimplemented }
func (op _mod) step(_ *mach) error     { return errUnimplemented }
func (op _div) step(_ *mach) error     { return errUnimplemented }
func (op _lt) step(_ *mach) error      { return errUnimplemented }
func (op _eq) step(_ *mach) error      { return errUnimplemented }
func (op label) step(_ *mach) error    { return errUnimplemented }
func (op push) step(_ *mach) error     { return errUnimplemented }
func (op fnz) step(_ *mach) error      { return errUnimplemented }
func (op jnz) step(_ *mach) error      { return errUnimplemented }
func (op jz) step(_ *mach) error       { return errUnimplemented }
func (op labelRef) step(_ *mach) error { return errUnimplemented }
func (op _halt) step(_ *mach) error    { return errUnimplemented }
func (op _hnz) step(_ *mach) error     { return errUnimplemented }
func (op _hz) step(_ *mach) error      { return errUnimplemented }
