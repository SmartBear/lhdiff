package main

import (
	"flag"
	"github.com/SmartBear/lhdiff"
	"io/ioutil"
)

func main() {
	compact := flag.Bool("compact", false, "Exclude identical lines from output")
	flag.Parse()
	leftFile := flag.Arg(0)
	rightFile := flag.Arg(1)
	left, _ := ioutil.ReadFile(leftFile)
	right, _ := ioutil.ReadFile(rightFile)
	linePairs, leftCount, newRightLines := lhdiff.Lhdiff(string(left), string(right), 4)
	lhdiff.PrintLinePairs(linePairs, leftCount, newRightLines, !*compact)
}
