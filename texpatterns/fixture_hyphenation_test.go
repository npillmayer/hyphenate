package texpatterns

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/npillmayer/hyphenate/texexceptions"
)

func mustLoadFixture(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "testdata", file))
	if err != nil {
		t.Fatalf("cannot read fixture %s: %v", file, err)
	}
	return data
}

func TestPatternsAndExceptionsLoadSeparately(t *testing.T) {
	src := `\hyphenation{
ta-ble
}`
	dict, err := LoadPatterns("split-api-test", strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if h := dict.HyphenationString("table"); h != "table" {
		t.Fatalf("without exceptions table should remain table, is %s", h)
	}
	texexceptions.LoadExceptions(dict, strings.NewReader(src))
	if h := dict.HyphenationString("table"); h != "ta-ble" {
		t.Fatalf("with exceptions table should be ta-ble, is %s", h)
	}
}

func TestUSPatternsFixture(t *testing.T) {
	data := mustLoadFixture(t, "hyph-en-us.tex")
	dict, err := LoadPatterns("hyph-en-us.tex", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	texexceptions.LoadExceptions(dict, bytes.NewReader(data))

	tests := []struct {
		word string
		want string
	}{
		{word: "hello", want: "hel-lo"},
		{word: "table", want: "ta-ble"},
		{word: "computer", want: "com-put-er"},
		{word: "algorithm", want: "al-go-rithm"},
		{word: "concatenation", want: "con-cate-na-tion"},
		{word: "quick", want: "quick"},
		{word: "king", want: "king"},
	}
	for _, tt := range tests {
		if got := dict.HyphenationString(tt.word); got != tt.want {
			t.Fatalf("hyphenation mismatch for %q: got %q, want %q", tt.word, got, tt.want)
		}
	}
}

func TestGermanPatternFixtureUmlauts(t *testing.T) {
	data := mustLoadFixture(t, "hyph-de-1996.tex")
	dict, err := LoadPatterns("hyph-de-1996.tex", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	texexceptions.LoadExceptions(dict, bytes.NewReader(data))

	tests := []struct {
		word string
		want string
	}{
		{word: "Mädchen", want: "Mäd-chen"},
		{word: "schönheit", want: "schön-heit"},
		{word: "frühling", want: "früh-ling"},
		{word: "häuser", want: "häu-ser"},
		{word: "öffentlichkeit", want: "öf-fent-lich-keit"},
		{word: "mäßig", want: "mä-ßig"},
		{word: "übergröße", want: "über-grö-ße"},
	}
	for _, tt := range tests {
		got := dict.HyphenationString(tt.word)
		if got != tt.want {
			t.Fatalf("hyphenation mismatch for %q: got %q, want %q", tt.word, got, tt.want)
		}
		if strings.ReplaceAll(got, "-", "") != tt.word {
			t.Fatalf("hyphenation corrupted original word %q -> %q", tt.word, got)
		}
	}
}

func TestUnicodeExceptionSplit(t *testing.T) {
	dict, err := LoadPatterns("unicode-test", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	texexceptions.LoadExceptions(dict, strings.NewReader(`\hyphenation{
fü-rung
schön-heit
}`))
	if h := dict.HyphenationString("fürung"); h != "fü-rung" {
		t.Fatalf("fürung should be fü-rung, is %s", h)
	}
	if h := dict.HyphenationString("schönheit"); h != "schön-heit" {
		t.Fatalf("schönheit should be schön-heit, is %s", h)
	}
}
