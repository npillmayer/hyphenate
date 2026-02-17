package texexceptions

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/npillmayer/hyphenate"
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

func TestUnicodeExceptionSplit(t *testing.T) {
	dict, err := hyphenate.LoadPatterns("unicode-test", emptyPatternReader{})
	if err != nil {
		t.Fatal(err)
	}
	LoadExceptions(dict, strings.NewReader(`\hyphenation{
füh-rung
schön-heit
}`))
	if h := dict.HyphenationString("führung"); h != "füh-rung" { // from exceptions
		t.Fatalf("führung should be füh-rung, is %s", h)
	}
	if h := dict.HyphenationString("schönheit"); h != "schön-heit" { // from exceptions
		t.Fatalf("schönheit should be schön-heit, is %s", h)
	}
}

func TestUSPatternsFixture(t *testing.T) {
	dict, err := hyphenate.LoadPatterns("none", emptyPatternReader{})
	if err != nil {
		t.Fatal(err)
	}
	data := mustLoadFixture(t, "hyph-en-us.tex")
	LoadExceptions(dict, bytes.NewReader(data))

	tests := []struct {
		word string
		want string
	}{
		{word: "table", want: "ta-ble"},
		{word: "present", want: "present"},
		{word: "reformation", want: "ref-or-ma-tion"},
	}
	for _, tt := range tests {
		if got := dict.HyphenationString(tt.word); got != tt.want {
			t.Fatalf("hyphenation mismatch for %q: got %q, want %q", tt.word, got, tt.want)
		}
	}
}

// ----------------------------------------------------------------------

func mustLoadFixture(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", file))
	if err != nil {
		t.Fatalf("cannot read fixture %s: %v", file, err)
	}
	return data
}

type emptyPatternReader struct{}

func (r emptyPatternReader) Next() (sequence []rune, weights []int, err error) {
	return nil, nil, io.EOF
}
