package main

import "fmt"

type debugWatcher struct{}

func (wat debugWatcher) emitted(srch searcher, parent, child *solution) {
	fmt.Printf("+++ %v %v", srch.frontierSize(), child)
	if parent != nil {
		fmt.Printf(" parent %v @%v", parent.steps[parent.stepi], parent.stepi)
	}
	fmt.Printf("\n")
}

func (wat debugWatcher) beforeStep(srch searcher, sol *solution) {
	fmt.Printf(">>> %v\n", sol)
}

func (wat debugWatcher) stepped(srch searcher, sol *solution) {
	fmt.Printf("... %v\n", sol)
	if _, ok := sol.steps[sol.stepi-1].(storeStep); ok {
		fmt.Printf("... %s\n", sol.letterMapping())
	} else if _, ok := sol.steps[sol.stepi-1].(forkUntilStep); ok {
		fmt.Printf("... len(frontier) == %v\n", srch.frontierSize())
	}
}
