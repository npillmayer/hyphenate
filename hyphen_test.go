package hyphenate

import (
	"os"
	"strings"
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

var germanExceptionsDict, germanPatternsDict, usDict *Dictionary

//germanDict = LoadPatterns(gconf.GetString("etc-dir") + "/pattern/hyph-de-1996.tex")
//usDict = LoadPatterns(gconf.GetString("etc-dir") + "/pattern/hyph-en-us.tex")
//usDict = LoadPatterns("/Users/npi/prg/go/gotype/etc/hyph-en-us.tex")

func init() {
	germanExceptionsDict = LoadPatterns("de-test", strings.NewReader(`\hyphenation{
Aus-nah-me
}`))
	f, err := os.Open("testdata/hyph-en-us.tex")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	usDict = LoadPatterns("hyph-en-us.tex", f)
	df, err := os.Open("testdata/hyph-de-1996.tex")
	if err != nil {
		panic(err)
	}
	defer df.Close()
	germanPatternsDict = LoadPatterns("hyph-de-1996.tex", df)
}

func TestDEPatterns(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "hyphenate")
	defer teardown()
	//
	//fmt.Printf("Ausnahme = %s\n", dict.HyphenationString("Ausnahme"))
	h := germanExceptionsDict.HyphenationString("Ausnahme")
	if h != "Aus-nah-me" {
		t.Fail()
	}
}

func TestDEPatterns2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "hyphenate")
	defer teardown()
	//
	s := germanExceptionsDict.Hyphenate("Ausnahme")
	t.Logf("Ausnahme = %v (%d)\n", s, len(s))
	if len(s) != 3 || s[0] != "Aus" {
		t.Fail()
	}
}

func TestUSPatterns(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "hyphenate")
	defer teardown()
	//
	h := usDict.HyphenationString("hello")
	if h != "hel-lo" {
		t.Logf("hello should be hel-lo, is %s", h)
		t.Fail()
	}
	h = usDict.HyphenationString("table") // exception dictionary
	if h != "ta-ble" {
		t.Logf("table should be ta-ble, is %s", h)
		t.Fail()
	}
	h = usDict.HyphenationString("computer")
	if h != "com-put-er" {
		t.Logf("computer should be com-put-er, is %s", h)
		t.Fail()
	}
	h = usDict.HyphenationString("algorithm")
	if h != "al-go-rithm" {
		t.Logf("algorithm should be al-go-rithm, is %s", h)
		t.Fail()
	}
	h = usDict.HyphenationString("concatenation")
	if h != "con-cate-na-tion" {
		t.Logf("concatenation should be con-cate-na-tion, is %s", h)
		t.Fail()
	}
	h = usDict.HyphenationString("quick")
	if h != "quick" {
		t.Logf("quick should be quick, is %s", h)
		t.Fail()
	}
	h = usDict.HyphenationString("king")
	if h != "king" {
		t.Logf("king should be king, is %s", h)
		t.Fail()
	}
}

func TestUnicodeExceptionSplit(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "hyphenate")
	defer teardown()
	//
	dict := LoadPatterns("unicode-test", strings.NewReader(`\hyphenation{
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

func TestUnicodePatternSplit(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "hyphenate")
	defer teardown()
	//
	dict := LoadPatterns("unicode-pattern-test", strings.NewReader(`
fü1r
`))
	if h := dict.HyphenationString("fürung"); h != "fü-rung" {
		t.Fatalf("fürung should be fü-rung, is %s", h)
	}
}

func TestPatternTrieStats(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "hyphenate")
	defer teardown()
	//
	backend, used, total, maxStateID, fill := usDict.PatternTrieStats()
	t.Logf("pattern trie stats: backend=%s used=%d total=%d fill=%.4f maxStateID=%d",
		backend, used, total, fill, maxStateID)
	if backend != "dat" {
		t.Fatalf("expected dat backend, got %s", backend)
	}
	if used <= 0 || total <= 0 {
		t.Fatalf("expected positive slot counts, got used=%d total=%d", used, total)
	}
	if maxStateID <= 0 {
		t.Fatalf("expected positive maxStateID, got %d", maxStateID)
	}
	if fill <= 0 || fill > 1 {
		t.Fatalf("expected fill ratio in (0,1], got %f", fill)
	}
}

func TestGermanPatternFixtureUmlauts(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "hyphenate")
	defer teardown()
	//
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
		got := germanPatternsDict.HyphenationString(tt.word)
		if got != tt.want {
			t.Fatalf("hyphenation mismatch for %q: got %q, want %q", tt.word, got, tt.want)
		}
		if strings.ReplaceAll(got, "-", "") != tt.word {
			t.Fatalf("hyphenation corrupted original word %q -> %q", tt.word, got)
		}
	}
}
