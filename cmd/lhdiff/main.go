package main

import (
	"flag"
	"github.com/aslakhellesoy/lhdiff"
	"io/ioutil"
)

func main() {
	omit := flag.Bool("omit", false, "Omit lines with same line numbers from output")
	flag.Parse()
	leftFile := flag.Arg(0)
	rightFile := flag.Arg(1)
	left, _ := ioutil.ReadFile(leftFile)
	right, _ := ioutil.ReadFile(rightFile)
	linePairs, leftCount, newRightLines := lhdiff.Lhdiff(string(left), string(right), 4)
	lhdiff.PrintLinePairs(linePairs, leftCount, newRightLines, *omit)
}
