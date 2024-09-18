package jwasm_test

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/jcklie/jwasm"
)

func TestUnsigned(t *testing.T) {
	for _, test := range []struct {
		Hex           string
		ValueAsString string
	}{
		{"00", "0"},
		{"07", "7"},
		{"7F", "127"},
		{"E58E26", "624485"},
		{"80897A", "2000000"},
		{"808098F4E9B5CA6A", "60000000000000000"},
		{"EF9BAF8589CF959A92DEB7DE8A929EABB424", "24197857200151252728969465429440056815"},
	} {
		t.Run(test.Hex, func(t *testing.T) {
			expected, success := new(big.Int).SetString(test.ValueAsString, 10)
			if !success {
				t.Fatalf("Failed to parse value: [%s]", test.ValueAsString)
			}

			buf, err := hex.DecodeString(test.Hex)
			if err != nil {
				t.Fatal(err)
			}
			r := bytes.NewReader(buf)

			actual, err := jwasm.ReadUleb128(r)
			if err != nil {
				t.Fatal(err)
			}

			if expected.Cmp(actual) != 0 {
				t.Errorf("%s:\nexpected: %s\nactual: %s", test.Hex, expected, actual)
			}
			if r.Len() != 0 {
				t.Error()
			}
		})
	}
}
