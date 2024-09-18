package jwasm

import (
	"fmt"
	"io"
)

func parseName(r io.Reader) (string, error) {
	// https://webassembly.github.io/spec/core/binary/values.html#binary-name
	//
	// Names are encoded as a vector of bytes containing the Unicode (Section 3.9)
	// UTF-8 encoding of the name’s character sequence.
	size, err := ReadUint32(r)
	if err != nil {
		return "", fmt.Errorf("reading vector size failed: %w", err)
	}

	bytes := make([]byte, size)
	bytesRead, err := r.Read(bytes)

	if bytesRead != int(size) {
		return "", fmt.Errorf("reading vector failed, wanted [%d] bytes, got only [%d] bytes", size, bytesRead)
	}

	if err != nil {
		return "", fmt.Errorf("reading vector data failed: %w", err)
	}

	return string(bytes), nil
}
