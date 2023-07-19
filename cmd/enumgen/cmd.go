package main

import (
	"github.com/daichitakahashi/go-enum/cmd/enumgen/gen"
)

func main() {
	gen.Run("./user", "enum.gen.go")
}
