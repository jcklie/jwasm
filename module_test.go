package jwasm

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingCustomSection(t *testing.T) {
	data := []byte{0x04, 0x6E, 0x61, 0x6D, 0x65, 0x02, 0x01, 0x00}
	r := bytes.NewReader(data[:])

	section, err := parseCustomSection(r)

	if err != nil {
		t.Error(err)
	}

	assert.NotNil(t, section)
	assert.Equal(t, section.Name, "name", "section name should be 'name'")
	assert.Equal(t, section.Data, []byte{0x02, 0x01, 0x00}, "section data should be [2, 1, 0]")

}
