package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

// nolint
var (
	version = "snapshot"
	commit  = "snapshot"
	date    = "snapshot"
)

func main() {
	app := kingpin.New("lhdiff", "A Lightweight Hybrid Approach for Tracking Source Lines").Version(version).Author("aslakhellesoy")
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	}
}
