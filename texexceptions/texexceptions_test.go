package texexceptions

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	src := strings.NewReader(`\hyphenation{
ta-ble
schön-heit
}`)
	r := NewReader(src)
	word, positions, err := r.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if word != "table" {
		t.Fatalf("word mismatch: got %q", word)
	}
	if !reflect.DeepEqual(positions, []int{0, 0, 1, 0, 0}) {
		t.Fatalf("positions mismatch: %v", positions)
	}
	word, positions, err = r.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if word != "schönheit" {
		t.Fatalf("word mismatch: got %q", word)
	}
	if !reflect.DeepEqual(positions, []int{0, 0, 0, 0, 0, 1, 0, 0, 0}) {
		t.Fatalf("positions mismatch: %v", positions)
	}
	_, _, err = r.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}
