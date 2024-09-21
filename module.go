package jwasm

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Indices
//  https://webassembly.github.io/spec/core/binary/modules.html#indices

type typeIndex uint32
type functionIndex uint32
type tableIndex uint32
type memoryIndex uint32
type globalIndex uint32
type elementIndex uint32
type dataIndex uint32
type localIndex uint32
type labelIndex uint32

// Section
// https://webassembly.github.io/spec/core/binary/modules.html#sections

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
	case exportSectionId:
		return parseExportSection(limitReader)
	case codeSectionId:
		return parseCodeSection(limitReader)
	default:
		return nil, fmt.Errorf("reading of section with unknown id failed: %d", sectionId)
	}
}

// Custom Section

type CustomSection struct {
	Name string
	Data []byte
}

func (cs *CustomSection) section() {}

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

// Type Section

type TypeSection struct {
	FunctionTypes []FunctionType
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

func (cs *TypeSection) section() {}

// Function Section

type FunctionSection struct {
	typeIndices []uint32
}

func (cs *FunctionSection) section() {}

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

// Export Section

type ExportSection struct {
	exports []export
}

func (cs *ExportSection) section() {}

// https://webassembly.github.io/spec/core/syntax/modules.html#syntax-exportdesc
type export struct {
	name              string
	exportDescription exportDescription
}

// https://webassembly.github.io/spec/core/syntax/modules.html#syntax-exportdesc
type exportDescription interface {
	exportDescription()
}

type exportDescriptionFunc struct {
	functionIndex functionIndex
}

type exportDescriptionTable struct {
	tableIndex tableIndex
}

type exportDescriptionMem struct {
	memoryIndex memoryIndex
}

type exportDescriptionGlobal struct {
	globalIndex globalIndex
}

func (*exportDescriptionFunc) exportDescription()   {}
func (*exportDescriptionTable) exportDescription()  {}
func (*exportDescriptionMem) exportDescription()    {}
func (*exportDescriptionGlobal) exportDescription() {}

func (exportDescriptionFunc) String() string   { return "func" }
func (exportDescriptionTable) String() string  { return "table" }
func (exportDescriptionMem) String() string    { return "mem" }
func (exportDescriptionGlobal) String() string { return "global" }

func parseExportSection(r io.Reader) (*ExportSection, error) {
	// https://webassembly.github.io/spec/core/binary/modules.html#export-section
	//
	// The export section has the id 7. It decodes into a vector of exports that represent the
	// `exports` component of a module.

	numExports, err := ReadUint32(r)
	if err != nil {
		return nil, fmt.Errorf("reading vector size of export section failed: %w", err)
	}

	var exports []export
	for i := 0; i < int(numExports); i++ {
		name, err := parseName(r)
		if err != nil {
			return nil, fmt.Errorf("parsing name of export failed: %w", err)
		}

		var b byte
		err = binary.Read(r, binary.BigEndian, &b)
		if err != nil {
			return nil, fmt.Errorf("reading export description type byte failed: %w", err)
		}

		var x uint32
		err = binary.Read(r, binary.BigEndian, &b)
		if err != nil {
			return nil, fmt.Errorf("reading export description index failed: %w", err)
		}

		var exportDescription exportDescription
		switch b {
		case 0x00:
			exportDescription = &exportDescriptionFunc{functionIndex(x)}
		case 0x01:
			exportDescription = &exportDescriptionTable{tableIndex(x)}
		case 0x02:
			exportDescription = &exportDescriptionMem{memoryIndex(x)}
		case 0x03:
			exportDescription = &exportDescriptionGlobal{globalIndex(x)}
		}

		export := export{name, exportDescription}

		exports = append(exports, export)
	}

	return &ExportSection{exports}, nil
}

// Code Section

type CodeSection struct {
	functionCode []functionCode
}

func (cs *CodeSection) section() {}

// https://webassembly.github.io/spec/core/binary/modules.html#binary-func
type functionCode struct {
	locals map[ValueType]uint32
	body   []instruction
}

func parseCodeSection(r io.Reader) (*CodeSection, error) {
	// https://webassembly.github.io/spec/core/binary/modules.html#code-section
	//
	// The code section has the id 10. It decodes into a vector of code entries
	// that are pairs of value type vectors and expressions. They represent the
	// `locals` and `body` field of the functions in the `funcs` component of a
	// module. The `type` fields of the respective functions are encoded separately
	// in the function section.

	numEntries, err := ReadUint32(r)
	if err != nil {
		return nil, fmt.Errorf("reading vector size of code section failed: %w", err)
	}

	var result []functionCode
	for i := 0; i < int(numEntries); i++ {
		codeSize, err := ReadUint32(r)
		if err != nil {
			return nil, fmt.Errorf("reading vector size of function code failed: %w", err)
		}

		// Parse locals
		limitReader := io.LimitReader(r, int64(codeSize))
		numLocals, err := ReadUint32(limitReader)
		if err != nil {
			return nil, fmt.Errorf("reading vector size of function locals failed: %w", err)
		}

		locals := make(map[ValueType]uint32)
		for j := 0; j < int(numLocals); j++ {
			n, err := ReadUint32(limitReader)
			if err != nil {
				return nil, fmt.Errorf("reading function locals n failed: %w", err)
			}

			valueType, err := parseValueType(limitReader)
			if err != nil {
				return nil, fmt.Errorf("reading function locals value type failed: %w", err)
			}

			locals[valueType] += n
		}

		instructions, err := parseInstructions(limitReader)
		if err != nil {
			return nil, fmt.Errorf("reading function body failed: %w", err)
		}

		functionCode := functionCode{locals, instructions}
		result = append(result, functionCode)
	}

	return &CodeSection{result}, nil
}
