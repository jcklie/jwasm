package jwasm

import (
	"encoding/binary"
	"fmt"
	"io"
)

type instruction interface {
	// Call(vm *VM)
	instruction()
}

// Variable Instructions
// https://webassembly.github.io/spec/core/binary/instructions.html#variable-instructions

type localGet struct{ x localIndex }

func (*localGet) instruction() {}

func parseLocalGet(r io.Reader) (*localGet, error) {
	x, err := ReadUint32(r)
	if err != nil {
		return nil, fmt.Errorf("reading x for LocalGet failed: %w", err)
	}
	return &localGet{localIndex(x)}, nil
}

// Numeric Instructions
// https://webassembly.github.io/spec/core/binary/instructions.html#numeric-instructions

type int32Add struct{}

func (*int32Add) instruction() {}

func parseInstructions(r io.Reader) ([]instruction, error) {
	// https://webassembly.github.io/spec/core/binary/instructions.html#instructions
	//
	// Instructions are encoded by opcodes. Each opcode is represented by a single
	// byte, and is followed by the instruction’s immediate arguments, where present.
	// The only exception are structured control instructions, which consist of several
	// opcodes bracketing their nested instruction sequences.

	var instructions []instruction

	for {
		var opcode byte
		err := binary.Read(r, binary.BigEndian, &opcode)

		if err != nil {
			return nil, fmt.Errorf("reading instruction byte failed: %w", err)
		}

		// https://webassembly.github.io/spec/core/binary/instructions.html#expressions
		if opcode == 0x0B {
			break
		}

		instruction, err := parseInstruction(r, opcode)

		if err != nil {
			return nil, fmt.Errorf("parsing instructions failed, unknown opcode: [%#X]", opcode)
		}

		instructions = append(instructions, instruction)
	}

	return instructions, nil
}

func parseInstruction(r io.Reader, opcode byte) (instruction, error) {

	switch opcode {
	// Variable Instructions
	case 0x20:
		return parseLocalGet(r)
	// Numeric Instructions
	case 0x6A:
		return &int32Add{}, nil
	default:
		return nil, fmt.Errorf("parsing instructions failed, unknown opcode: [%#X]", opcode)
	}
}
