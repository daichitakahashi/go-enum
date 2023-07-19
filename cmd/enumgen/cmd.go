package main

import (
	"flag"

	"github.com/daichitakahashi/go-enum/cmd/enumgen/gen"
)

var (
	wd  = flag.String("wd", ".", "working directory")
	out = flag.String("out", "enum.gen.go", "output file name")
)

func main() {
	flag.Parse()
	gen.Run(*wd, *out)
}
