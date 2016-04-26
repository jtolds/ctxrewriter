package main

import (
	"flag"
	"fmt"

	"github.com/jtolds/ctxrewriter"
)

var (
	inplaceFlag = flag.Bool("w", false,
		"if true, write to source file instead of stdout")
)

func main() {
	flag.Parse()
	for _, filename := range flag.Args() {
		err := ctxrewriter.ProcessFile(filename, *inplaceFlag)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
	}
}
