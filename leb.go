package jwasm

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
)

func ReadUint8(r io.Reader) (uint8, error) {
	value, err := ReadUleb128(r)
	if err != nil {
		return 0, fmt.Errorf("reading uleb128 for uint8 failed: %w", err)
	}

	if value.BitLen() > 8 {
		return 0, fmt.Errorf("uleb128 is too large for an uint8: %s", value)
	}

	return uint8(value.Uint64()), nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	value, err := ReadUleb128(r)
	if err != nil {
		return 0, fmt.Errorf("reading uleb128 for uint8 failed: %w", err)
	}

	if value.BitLen() > 32 {
		return 0, fmt.Errorf("uleb128 is too large for an uint8: %s", value)
	}

	return uint32(value.Uint64()), nil
}

func ReadUleb128(r io.Reader) (*big.Int, error) {
	result := new(big.Int)
	var bytesRead uint

	for {
		var b uint8

		err := binary.Read(r, binary.LittleEndian, &b)

		if err != nil {
			return nil, fmt.Errorf("reading uleb128 byte failed: %w", err)
		}

		value := new(big.Int)
		value.SetUint64(uint64(b & 0b01111111))
		value.Lsh(value, 7*bytesRead)
		result = result.Or(result, value)
		bytesRead += 1

		// If highest bit is not set, then we read the last byte for this LEB128
		isLast := (b & (0b10000000) >> 7) == 0

		if isLast {
			break
		}
	}

	return result, nil
}

// func ReadSleb128(r io.Reader) (uint, error) {
//
// }
