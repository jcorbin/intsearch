package main

import "fmt"

type debugWatcher struct {
	labelFor func(*solution) string
}

func (wat debugWatcher) emitted(srch searcher, child *solution) {
	fmt.Printf("+++ %v %v%s", srch.frontierSize(), child, wat.labelFor(child))
	if parent := srch.current(); parent != nil {
		fmt.Printf("\n... parent %v @%v%s", parent.steps[parent.stepi], parent.stepi, wat.labelFor(parent))
	}
	fmt.Printf("\n")
}

func (wat debugWatcher) beforeStep(srch searcher, sol *solution) {
	fmt.Printf(">>> %v%s\n", sol, wat.labelFor(sol))
}

func (wat debugWatcher) stepped(srch searcher, sol *solution) {
	fmt.Printf("... %v%s\n", sol, wat.labelFor(sol))
	if _, ok := sol.steps[sol.stepi-1].(storeStep); ok {
		fmt.Printf("... %s\n", sol.letterMapping())
	} else if isForkStep(sol.steps[sol.stepi-1]) {
		fmt.Printf("... len(frontier) == %v\n", srch.frontierSize())
	}
}
