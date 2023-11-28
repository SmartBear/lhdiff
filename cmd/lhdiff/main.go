package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oselvar/lhdiff"
)

func main() {
	compact := flag.Bool("compact", false, "Exclude identical lines from output")
	flag.Parse()
	leftFile := flag.Arg(0)
	rightFile := flag.Arg(1)
	left, _ := os.ReadFile(leftFile)
	right, _ := os.ReadFile(rightFile)
	mappings, err := lhdiff.Lhdiff(string(left), string(right), 4, !*compact)

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	err = lhdiff.PrintMappings(mappings)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
