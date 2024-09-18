package main

import (
	"os"

	"github.com/jcklie/jwasm"
)

func main() {
	file, err := os.Open("testdata/wat2wasm/simple.wasm")

	if err != nil {
		panic(err)
	}

	parser := jwasm.Parser{}
	err = parser.Parse(file)

	if err != nil {
		panic(err)
	}
}
