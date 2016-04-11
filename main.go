package main

import (
	"flag"
	"log"
)

func main() {
	flag.Parse()
	word1 := flag.Arg(0)
	if len(word1) == 0 {
		log.Fatalf("missing word1 argument")
	}
	word2 := flag.Arg(1)
	if len(word2) == 0 {
		log.Fatalf("missing word2 argument")
	}
	word3 := flag.Arg(2)
	if len(word3) == 0 {
		log.Fatalf("missing word3 argument")
	}

	var (
		prob problem
		gen  logGen
	)

	if err := prob.plan(word1, word2, word3, &gen); err != nil {
		log.Fatalf("plan failed: %v", err)
	}
}
