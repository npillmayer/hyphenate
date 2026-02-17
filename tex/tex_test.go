package tex

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func mustLoadFixture(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", file))
	if err != nil {
		t.Fatalf("cannot read fixture %s: %v", file, err)
	}
	return data
}

func TestLoadDictionaryUSFixture(t *testing.T) {
	data := mustLoadFixture(t, "hyph-en-us.tex")
	dict, err := LoadDictionary("hyph-en-us.tex", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		word string
		want string
	}{
		{word: "hello", want: "hel-lo"},
		{word: "table", want: "ta-ble"}, // comes from TeX exceptions
		{word: "computer", want: "com-put-er"},
	}
	for _, tt := range tests {
		if got := dict.HyphenationString(tt.word); got != tt.want {
			t.Fatalf("hyphenation mismatch for %q: got %q, want %q", tt.word, got, tt.want)
		}
	}
}

func TestLoadDictionaryGermanFixtureUmlauts(t *testing.T) {
	data := mustLoadFixture(t, "hyph-de-1996.tex")
	dict, err := LoadDictionary("hyph-de-1996.tex", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		word string
		want string
	}{
		{word: "Mädchen", want: "Mäd-chen"},
		{word: "übergröße", want: "über-grö-ße"},
	}
	for _, tt := range tests {
		if got := dict.HyphenationString(tt.word); got != tt.want {
			t.Fatalf("hyphenation mismatch for %q: got %q, want %q", tt.word, got, tt.want)
		}
	}
}
