package jwasm

import (
	"encoding/binary"
	"fmt"
	"io"
)

const WASM_BINARY_MAGIC uint32 = 0x0061736D
const WASM_BINARY_VERSION uint32 = 0x01000000

type Parser struct {
}

func (p *Parser) Parse(r io.Reader) error {
	// Read magic header
	var magic uint32
	err := binary.Read(r, binary.BigEndian, &magic)
	if err != nil {
		return fmt.Errorf("reading magic failed: %w", err)
	}

	if magic != WASM_BINARY_MAGIC {
		return fmt.Errorf("magic did not match, expected [0x%x] got, [0x%x]", WASM_BINARY_MAGIC, magic)
	}

	// Read version
	var version uint32
	err = binary.Read(r, binary.BigEndian, &version)
	if err != nil {
		return fmt.Errorf("reading version failed: %w", err)
	}

	if version != WASM_BINARY_VERSION {
		return fmt.Errorf("version did not match, expected [0x%x] got, [0x%x]", WASM_BINARY_VERSION, magic)
	}

	// Parse sections
	for {
		section, err := parseSection(r)

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		fmt.Printf("Section: %+v\n", section)
	}

	return nil
}
