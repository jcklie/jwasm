package jwasm

import (
	"encoding/binary"
	"fmt"
	"io"
)

type instructionType int

type SectionId byte

const (
	customSectionId    SectionId = 0
	typeSectionId      SectionId = 1
	importSectionId    SectionId = 2
	functionSectionId  SectionId = 3
	tableSectionId     SectionId = 4
	memorySectionId    SectionId = 5
	globalSectionId    SectionId = 6
	exportSectionId    SectionId = 7
	startSectionId     SectionId = 8
	elementSectionId   SectionId = 9
	codeSectionId      SectionId = 10
	dataSectionId      SectionId = 11
	dataCountSectionId SectionId = 12
)

type Section interface {
	section()
}

type CustomSection struct {
	Name string
	Data []byte
}

type TypeSection struct {
	FunctionTypes []FunctionType
}

type FunctionSection struct {
	typeIndices []uint32
}

func (cs *CustomSection) section()   {}
func (cs *TypeSection) section()     {}
func (cs *FunctionSection) section() {}

func parseSection(r io.Reader) (Section, error) {
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

	// Parse section size
	sectionSize, err := ReadUint32(r)
	if err != nil {
		return nil, fmt.Errorf("reading section size failed: %w", err)
	}

	limitReader := io.LimitReader(r, int64(sectionSize))

	// Parse section
	switch SectionId(sectionId) {
	case customSectionId:
		return parseCustomSection(limitReader)
	case typeSectionId:
		return parseTypeSection(limitReader)
	case functionSectionId:
		return parseFunctionSection(limitReader)
	default:
		return nil, fmt.Errorf("reading of section with unknown id failed: %d", sectionId)
	}
}

func parseCustomSection(r io.Reader) (*CustomSection, error) {
	// https://webassembly.github.io/spec/core/binary/modules.html#custom-section
	//
	// Custom sections have the id 0. They are intended to be used for debugging
	// information or third-party extensions, and are ignored by the WebAssembly
	// semantics. Their contents consist of a name further identifying the custom
	// section, followed by an uninterpreted sequence of bytes for custom use.

	// Parse section name
	sectionName, err := parseName(r)
	if err != nil {
		return nil, fmt.Errorf("reading custom section name failed: %w", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading custom section data failed: %w", err)
	}

	// TOOD: Parse well-known custom sections, see
	// https://webassembly.github.io/spec/core/appendix/custom.html

	return &CustomSection{sectionName, data}, nil
}

func parseTypeSection(r io.Reader) (*TypeSection, error) {
	// https://webassembly.github.io/spec/core/binary/modules.html#type-section
	//
	// The type section decodes into a vector of function types that represent the
	// component types of a module.

	numTypes, err := ReadUint32(r)
	if err != nil {
		return nil, fmt.Errorf("reading vector size of function types failed: %w", err)
	}

	var functionTypes []FunctionType
	for i := 0; i < int(numTypes); i++ {
		functionType, err := parseFunctionType(r)

		if err != nil {
			return nil, fmt.Errorf("reading function type failed: %w", err)
		}

		functionTypes = append(functionTypes, functionType)
	}

	result := new(TypeSection)
	result.FunctionTypes = functionTypes
	return result, nil
}

func parseFunctionSection(r io.Reader) (*FunctionSection, error) {
	// https://webassembly.github.io/spec/core/binary/modules.html#function-section

	// The function section has the id 3. It decodes into a vector of type indices that represent
	// the `type` fields of the functions in the `funcs` component of a module. The `locals` and
	// and `body` fields of the respective functions are encoded separately in the code section.

	numIndices, err := ReadUint32(r)
	if err != nil {
		return nil, fmt.Errorf("reading vector size of function section failed: %w", err)
	}

	var typeIndices []uint32
	for i := 0; i < int(numIndices); i++ {
		typeIdx, err := ReadUint32(r)
		if err != nil {
			return nil, fmt.Errorf("reading type index in function section failed: %w", err)
		}

		typeIndices = append(typeIndices, typeIdx)
	}

	return &FunctionSection{typeIndices}, nil
}
