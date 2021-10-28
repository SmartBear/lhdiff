package main

import (
	"flag"
	"github.com/aslakhellesoy/lhdiff"
	"io/ioutil"
)

func main() {
	leftFile := flag.Arg(0)
	rightFile := flag.Arg(1)
	//app.Parse(os.Args[1:])
	left, _ := ioutil.ReadFile(leftFile)
	right, _ := ioutil.ReadFile(rightFile)
	lhdiff.Lhdiff(string(left), string(right))
}
