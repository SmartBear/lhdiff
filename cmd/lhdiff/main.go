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
	lhdiff.PrintLhdiff(string(left), string(right), false)
}
