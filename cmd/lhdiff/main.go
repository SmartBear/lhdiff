package main

import (
	"flag"
	"github.com/aslakhellesoy/lhdiff"
	"io/ioutil"
)

func main() {
	flag.Parse()
	leftFile := flag.Arg(0)
	rightFile := flag.Arg(1)
	left, _ := ioutil.ReadFile(leftFile)
	right, _ := ioutil.ReadFile(rightFile)
	linePairs := lhdiff.Lhdiff(string(left), string(right), 4)
	lhdiff.PrintLinePairs(linePairs, false)
}
