/*
Package main implements a utility that resolves or "sessionizes" debug from
"github.com/jcorbin/intsearch/opcode".TinyMachine.RunAll.

Here's an example of the output that we're interested in:

	0xc820094c80 RUN TinyMachine(@0000 h=false t=false r1=0 r2=0 r3=0 r4=0 r5=0)
	0xc820094dc0 ALLOC <- 0xc820094c80
	0xc820094dc0 PUSH
	0xc820094f00 ALLOC <- 0xc820094c80
	0xc820094f00 PUSH
	0xc820094c80 PUT
	0xc820094dc0 SHIFT
	0xc820094dc0 RUN TinyMachine((11/61184) @004e h=false t=false r1=0 r2=0 r3=0 r4=0 r5=0)
	0xc820094c80 REUSE <- 0xc820094dc0
	0xc820094c80 PUSH
	...
	0xc820095cc0 PUSH
	0xc820095a40 PUT
	0xc820094f00 SHIFT
	0xc820094f00 RUN TinyMachine((153/61184) @0bbd h=false t=true r1=0 r2=16 r3=8 r4=0 r5=7)
	0xc820094f00 PUT
	0xc820095400 SHIFT
	0xc820095400 RUN TinyMachine((190/61184) @10f8 h=false t=true r1=3 r2=6 r3=7 r4=1 r5=8)
	0xc820095a40 REUSE <- 0xc820095400
	0xc820095a40 PUSH
	0xc820095400 PUT
	0xc82010e140 SHIFT
	0xc82010e140 RUN TinyMachine((190/61184) @10f8 h=false t=true r1=4 r2=8 r3=6 r4=1 r5=8)
	0xc820095400 REUSE <- 0xc82010e140
	0xc820095400 PUSH
	0xc82010e140 SOL_1 TinyMachine((342/61184) @153c h=true t=true r1=6 r2=0 r3=0 r4=1 r5=5)
	...

So the question that this tool answers is "what was the full history of each solution?"

To get there, we first resolve each pointer's session; each session starts with
an ALLOC or REUSE line, and ends with a PUT line.

Parent linkage is recovered from each ALLOC and REUSE line, and a chain of
sessions built.

For each resolved session, we then print output like:

	[0xc820094dc0 0xc820094c80 0xc820094f00 0xc820094dc0 0xc820094c80 0xc820094f00 0xc820094dc0 0xc820094c80 0xc820095400 0xc820094f00 0xc8200952c0 0xc82010e000 0xc820095040 0xc820095180]:
	- ALLOC(0xc820094dc0): [<- 0xc820094c80]
	- PUSH(0xc820094dc0)
	- SHIFT(0xc820094dc0)
	- RUN(0xc820094dc0): [TinyMachine((11/61184) @004e h=false t=false r1=0 r2=0 r3=0 r4=0 r5=0)]
	- PUT(0xc820094dc0)
	  - REUSE(0xc820094c80): [<- 0xc820094dc0]
	  - PUSH(0xc820094c80)
	  - SHIFT(0xc820094c80)
	  - RUN(0xc820094c80): [TinyMachine((16/61184) @0066 h=false t=true r1=2 r2=4 r3=0 r4=0 r5=0)]
	  - PUT(0xc820094c80)
		- REUSE(0xc820094f00): [<- 0xc820094c80]
		- PUSH(0xc820094f00)
		- SHIFT(0xc820094f00)
		- RUN(0xc820094f00): [TinyMachine((23/61184) @0066 h=false t=true r1=3 r2=6 r3=0 r4=0 r5=0)]
		- PUT(0xc820094f00)
		  - REUSE(0xc820094dc0): [<- 0xc820094f00]
		  - PUSH(0xc820094dc0)
		  - SHIFT(0xc820094dc0)
		  - RUN(0xc820094dc0): [TinyMachine((30/61184) @0066 h=false t=true r1=4 r2=8 r3=0 r4=0 r5=0)]
		  - PUT(0xc820094dc0)
			- REUSE(0xc820094c80): [<- 0xc820094dc0]
			- PUSH(0xc820094c80)
	...

*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type act int

const (
	actAlloc = act(iota)
	actPush
	actPut
	actReuse
	actRun
	actShift
	actSol
	actInvalid = act(-1)
)

func (a act) String() string {
	switch a {
	case actAlloc:
		return "ALLOC"
	case actPush:
		return "PUSH"
	case actPut:
		return "PUT"
	case actReuse:
		return "REUSE"
	case actRun:
		return "RUN"
	case actShift:
		return "SHIFT"
	case actSol:
		return "SOL"
	}
	return "INVALID"
}

func actionCode(s string) act {
	switch s {
	case "ALLOC":
		return actAlloc
	case "PUSH":
		return actPush
	case "PUT":
		return actPut
	case "REUSE":
		return actReuse
	case "RUN":
		return actRun
	case "SHIFT":
		return actShift
	}

	if strings.HasPrefix(s, "SOL_") {
		return actSol
	}

	log.Printf("unsupported action %q", s)
	return actInvalid
}

type event struct {
	id    string
	act   act
	parts []string
}

func (ev *event) String() string {
	if len(ev.parts) > 1 {
		return fmt.Sprintf("%v(%s): %v", ev.act, ev.id, ev.parts)
	}
	return fmt.Sprintf("%v(%s)", ev.act, ev.id)
}

type session struct {
	parent *session
	id     string
	events []*event
}

func (se *session) parents() []*session {
	var ss []*session
	ps := se
	for ; ps.parent != nil; ps = ps.parent {
		ss = append(ss, ps)
	}
	i := 0
	j := len(ss) - 1
	for i < j {
		ss[i], ss[j] = ss[j], ss[i]
		i++
		j--
	}
	return ss
}

func (se *session) String() string {
	ss := se.parents()
	ids := make([]string, len(ss))
	var evs []string // TODO: pre-flight sum cap
	for i, s := range ss {
		ids[i] = s.id
		prefix := strings.Repeat("  ", i)
		for _, ev := range s.events {
			evs = append(evs, fmt.Sprintf("%s- %v", prefix, ev))
		}
	}
	return fmt.Sprintf("%v:\n%s",
		ids, // strings.Join(ids, " > ")
		strings.Join(evs, "\n"))
}

func eventFromLine(line string) *event {
	if !strings.HasPrefix(line, "0x") {
		return nil
	}
	parts := strings.Fields(line)
	act := actionCode(parts[1])
	if act == actInvalid {
		return nil
	}
	return &event{
		id:    parts[0],
		act:   act,
		parts: parts[2:],
	}
}

func scanEvents(r io.Reader, each func(ev *event)) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		ev := eventFromLine(line)
		if ev == nil {
			continue
		}
		each(ev)
	}
	return sc.Err()
}

func scanSolutions(r io.Reader, each func(se *session)) error {
	sess := make(map[string]*session)
	return scanEvents(r, func(ev *event) {
		se := sess[ev.id]
		if se == nil {
			se = &session{id: ev.id}
			sess[ev.id] = se
		}
		se.events = append(se.events, ev)
		switch ev.act {
		case actAlloc:
			fallthrough
		case actReuse:
			se.parent = sess[ev.parts[1]]
		case actPut:
			delete(sess, ev.id)
		case actSol:
			each(se)
		}
	})
}

func main() {
	if err := scanSolutions(os.Stdin, func(se *session) {
		fmt.Printf("%v\n", se)
	}); err != nil {
		log.Fatal(err)
	}
}
