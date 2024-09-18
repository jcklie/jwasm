package jwasm

import (
	"encoding/binary"
	"fmt"
	"io"
)

const WASM_BINARY_MAGIC uint32 = 0x0061736D
const WASM_BINARY_VERSION uint32 = 0x01000000

type instructionType int

type Parser struct {
}

type CustomSection struct {
	Name string
	Data []byte
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

	for {
		// Parse section
		_, err := parseSection(r)

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func parseSection(r io.Reader) (*CustomSection, error) {
	// https://webassembly.github.io/spec/core/binary/modules.html#sections
	// Each section consists of
	// - a one-byte section id,
	// - the size of the contents, in bytes,
	// - the actual contents, whose structure is dependent on the section id.

	// Parse section id
	var sectionId byte
	err := binary.Read(r, binary.BigEndian, &sectionId)

	if err != nil {
		if err == io.EOF {
			return nil, err
		} else {
			return nil, fmt.Errorf("reading section id failed: %w", err)
		}
	}
	fmt.Printf("Section id=[%d]\n", sectionId)

	// Parse section size
	sectionSize, err := ReadUint32(r)
	if err != nil {
		return nil, fmt.Errorf("reading section size failed: %w", err)
	}
	fmt.Printf("Section size=[%d]\n", sectionSize)

	// Parse section name
	sectionName, err := parseName(r)
	if err != nil {
		return nil, fmt.Errorf("reading section name failed: %w", err)
	}
	fmt.Printf("Section name=[%s]\n", sectionName)

	return nil, nil
}

func parseName(r io.Reader) (string, error) {
	bytes, err := parseVector(r)

	if err != nil {
		return "", fmt.Errorf("reading name failed: %w", err)
	}

	return string(bytes), nil
}

func parseVector(r io.Reader) ([]byte, error) {
	size, err := ReadUint8(r)
	if err != nil {
		return nil, fmt.Errorf("reading vector size failed: %w", err)
	}

	bytes := make([]byte, size)
	bytesRead, err := r.Read(bytes)

	if bytesRead != int(size) {
		return nil, fmt.Errorf("reading vector failed, wanted [%d] bytes, got only [%d] bytes", size, bytesRead)
	}

	if err != nil {
		return nil, fmt.Errorf("reading vector data failed: %w", err)
	}

	return bytes, nil
}

type Interpreter struct {
}

type VM struct {
}

type Instruction interface {
	Call(vm *VM)
	Type() instructionType
}
