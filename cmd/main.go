package main

import (
	"flag"
	"os"
	"path"

	"github.com/jcklie/jwasm"
)

var strFlag = flag.String("f", "<default>", "input file name")

func main() {
	flag.Parse()

	file, err := os.Open(path.Join("testdata/wat2wasm/", *strFlag))

	if err != nil {
		panic(err)
	}

	parser := jwasm.Parser{}
	err = parser.Parse(file)

	if err != nil {
		panic(err)
	}
}
