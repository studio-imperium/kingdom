package maps

import (
	"bytes"
	"testing"
)

func TestLoadRejectsBrokenMaps(t *testing.T) {
	for _, data := range [][]byte{nil, {0, 0}, {1, 0}, {1, 0, 0, 0, 0, 0, 10, 1, 1, 0}} {
		if _, err := Load(bytes.NewReader(data)); err == nil {
			t.Fatalf("accepted malformed map %v", data)
		}
	}
}
