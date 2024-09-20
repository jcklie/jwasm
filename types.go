package jwasm

import (
	"encoding/binary"
	"fmt"
	"io"
)

// https://webassembly.github.io/spec/core/syntax/types.html#number-types
type numberType struct {
	code numberTypeCode
	name string
}

func (*numberType) valueType()        {}
func (nt *numberType) String() string { return nt.name }

type numberTypeCode byte

const (
	NumberTypeI32 numberTypeCode = 0x7F
	NumberTypeI64 numberTypeCode = 0x7E
	NumberTypeF32 numberTypeCode = 0x7D
	NumberTypeF64 numberTypeCode = 0x7C
)

// https://webassembly.github.io/spec/core/syntax/types.html#vector-types
type vectorType struct {
	code vectorTypeCode
	name string
}

func (*vectorType) valueType()        {}
func (vt *vectorType) String() string { return vt.name }

type vectorTypeCode byte

const VectorTypeV128 vectorTypeCode = 0x7B

// https://webassembly.github.io/spec/core/syntax/types.html#reference-types
type referenceType struct {
	code referenceTypeCode
	name string
}

func (*referenceType) valueType()        {}
func (rt *referenceType) String() string { return rt.name }

type referenceTypeCode byte

const (
	ReferenceTypeFuncRef   referenceTypeCode = 0x70
	ReferenceTypeExternRef referenceTypeCode = 0x6F
)

// https://webassembly.github.io/spec/core/syntax/types.html#value-types
type ValueType interface {
	valueType()
}

func parseValueType(r io.Reader) (ValueType, error) {
	// https://webassembly.github.io/spec/core/binary/types.html#value-types
	//
	// Value types are encoded with their respective encoding as a number type, vector type, or reference type.

	var b byte
	err := binary.Read(r, binary.BigEndian, &b)
	if err != nil {
		return nil, fmt.Errorf("reading value type byte failed: %w", err)
	}

	switch b {
	// Number Type
	case byte(NumberTypeI32):
		return &numberType{NumberTypeI32, "i32"}, nil
	case byte(NumberTypeI64):
		return &numberType{NumberTypeI64, "i64"}, nil
	case byte(NumberTypeF32):
		return &numberType{NumberTypeF32, "f32"}, nil
	case byte(NumberTypeF64):
		return &numberType{NumberTypeF64, "f64"}, nil
	// Vector Type
	case byte(VectorTypeV128):
		return &vectorType{VectorTypeV128, "v128"}, nil
	// Reference Type
	case byte(ReferenceTypeFuncRef):
		return &referenceType{ReferenceTypeFuncRef, "funcref"}, nil
	case byte(ReferenceTypeExternRef):
		return &referenceType{ReferenceTypeExternRef, "externref"}, nil
	default:
		return nil, fmt.Errorf("reading value type failed, unknown code [0x%x]", b)
	}
}

// https://webassembly.github.io/spec/core/binary/types.html#result-types
type ResultType = []ValueType

func parseResultType(r io.Reader) (ResultType, error) {
	// https://webassembly.github.io/spec/core/binary/types.html#result-types
	//
	// Result types are encoded by the respective vectors of value types.

	// Read number of value types
	size, err := ReadUint32(r)
	if err != nil {
		return ResultType{}, fmt.Errorf("reading result type vector length failed: %w", err)
	}

	var resultType []ValueType

	for i := 0; i < int(size); i++ {
		valueType, err := parseValueType(r)

		if err != nil {
			return ResultType{}, fmt.Errorf("parsing value type for result type failed: %w", err)
		}

		resultType = append(resultType, valueType)
	}

	return resultType, nil
}

// https://webassembly.github.io/spec/core/binary/types.html#function-types
type FunctionType struct {
	ParameterTypes ResultType
	ResultTypes    ResultType
}

func parseFunctionType(r io.Reader) (FunctionType, error) {
	// https://webassembly.github.io/spec/core/binary/types.html#function-types
	//
	// Function types are encoded by the byte 0x60  followed by the respective
	// vectors of parameter and result types.

	// Read function type header
	var header byte
	err := binary.Read(r, binary.BigEndian, &header)
	if err != nil {
		return FunctionType{}, fmt.Errorf("reading function type header failed: %w", err)
	}

	if header != 0x60 {
		return FunctionType{}, fmt.Errorf("function type header wrong, expected [0x60], got [0x%x]", header)
	}

	rt1, err := parseResultType(r)
	if err != nil {
		return FunctionType{}, fmt.Errorf("parsing function type failed for rt1: %w", err)
	}

	rt2, err := parseResultType(r)
	if err != nil {
		return FunctionType{}, fmt.Errorf("parsing function type failed for rt2: %w", err)
	}

	return FunctionType{rt1, rt2}, nil
}
