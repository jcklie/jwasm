package jwasm

import "io"

func parseFunctionType(r io.Reader) (*TypeSection, error) {
	// https://webassembly.github.io/spec/core/binary/types.html#function-types
	//
	// Function types are encoded by the byte 0x60  followed by the respective
	// vectors of parameter and result types.
	return nil, nil
}
