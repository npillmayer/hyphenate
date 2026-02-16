package hyphenate

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	compacttrie "github.com/npillmayer/hyphenate/trie"
)

const (
	tinyTrieSize        = 15017 // must be prime
	tinyTrieCategoryCnt = 27    // dot + lowercase latin letters
)

func encodeTrieKey(s string) ([]byte, bool) {
	key := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '.':
			key = append(key, 1)
		case c >= 'a' && c <= 'z':
			key = append(key, c-'a'+2)
		default:
			return nil, false
		}
	}
	return key, true
}

func decodePatternLine(line string) (string, []int) {
	var pattern string
	var positions []int
	var wasdigit bool
	for _, char := range line {
		if unicode.IsDigit(char) {
			d, _ := strconv.Atoi(string(char))
			positions = append(positions, d)
			wasdigit = true
		} else {
			pattern = pattern + string(char)
			if wasdigit {
				wasdigit = false
			} else {
				positions = append(positions, 0)
			}
		}
	}
	return pattern, positions
}

func maxPackedEntriesFromPatternData(data []byte) uint8 {
	maxPacked := 0
	inExceptions := false
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if inExceptions {
			if strings.HasPrefix(line, "}") {
				inExceptions = false
			}
			continue
		}
		if strings.HasPrefix(line, "\\hyphenation{") {
			inExceptions = true
			continue
		}
		if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "\\") ||
			line == "" || strings.HasPrefix(line, "}") {
			continue
		}
		_, positions := decodePatternLine(line)
		nonzero := 0
		for _, v := range positions {
			if v != 0 {
				nonzero++
			}
		}
		if nonzero > maxPacked {
			maxPacked = nonzero
		}
	}
	if maxPacked > 16 {
		maxPacked = 16
	}
	return uint8(maxPacked)
}

// Dictionary is a loaded hyphenation dictionary.
//
// A dictionary contains:
//   - pattern rules (compiled into a TinyHashTrie + compact metadata store)
//   - explicit hyphenation exceptions from \hyphenation{...} blocks.
type Dictionary struct {
	exceptions map[string][]int // e.g., "computer" => [3,5] = "com-pu-ter"
	patterns   *compacttrie.TinyHashTrie
	patternsV  *patternStore // compact metadata vectors by pattern id
	Identifier string        // Identifies the dictionary
}

// LoadPatterns parses TeX hyphenation data and returns a ready-to-use dictionary.
//
// Patterns are enclosed in between
//
//	\patterns{ % some comment
//	 ...
//	.wil5i
//	.ye4
//	4ab.
//	a5bal
//	a5ban
//	abe2
//	 ...
//	}
//
// Odd numbers stand for possible discretionary breakpoints, even numbers forbid
// hyphenation. Digits belong to the character immediately after them, i.e.,
//
//	"a5ban" => (a)(5b)(a)(n) => positions["aban"] = [0,5,0,0].
//
// The loader uses a two-pass build:
//  1. pre-pass to determine compact payload width for metadata packing
//  2. trie build + freeze + metadata binding to final trie positions
func LoadPatterns(name string, reader io.Reader) *Dictionary {
	data, err := io.ReadAll(reader)
	if err != nil {
		panic(fmt.Sprintf("cannot read pattern source %q: %v", name, err))
	}
	maxPacked := maxPackedEntriesFromPatternData(data)
	patterns, err := compacttrie.NewTinyHashTrie(tinyTrieSize, tinyTrieCategoryCnt)
	if err != nil {
		panic(fmt.Sprintf("cannot initialize tiny trie: %v", err))
	}
	type pendingPattern struct {
		key       []byte
		positions []int
	}
	var pending []pendingPattern
	dict := &Dictionary{
		exceptions: make(map[string][]int),
		patterns:   patterns,
		patternsV:  newPatternStore(maxPacked),
		Identifier: fmt.Sprintf("patterns: %s", name),
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() { // internally, it advances token based on sperator
		line := scanner.Text()
		if strings.HasPrefix(line, "\\message{") { // extract patterns identifier
			dict.Identifier = line[9 : len(line)-1]
			//fmt.Println(dict.Identifier)
		} else if strings.HasPrefix(line, "\\hyphenation{") {
			dict.readExceptions(scanner)
		} else if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "\\") ||
			line == "" || strings.HasPrefix(line, "}") {
			// ignore comments, TeX commands, etc.
		} else { // read and decode a pattern: ".ab1a" "abe4l3in", ...
			pattern, positions := decodePatternLine(line)
			//fmt.Printf("pattern '%s'\thas positions %v\n", pattern, positions)
			key, ok := encodeTrieKey(pattern)
			if !ok {
				continue
			}
			if dict.patterns.AllocPositionForWord(key) == 0 {
				panic(fmt.Sprintf("could not allocate trie position for pattern %q", pattern))
			}
			pending = append(pending, pendingPattern{key: key, positions: positions})
		}
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Sprintf("error scanning pattern source %q: %v", name, err))
	}
	dict.patterns.Freeze()
	for _, p := range pending {
		patternID := dict.patterns.AllocPositionForWord(p.key)
		if patternID == 0 {
			panic("could not resolve trie position after freeze")
		}
		if err := dict.patternsV.Put(patternID, p.positions); err != nil {
			panic(fmt.Sprintf("cannot store pattern metadata at %d: %v", patternID, err))
		}
	}
	return dict
}

// readExceptions reads exceptions from a pattern file. Exceptions are denoted as
//
//	ex-cep-tion
//	ta-ble
//
// and so on, a single word per line. Exceptions are enclosed in
//
//	\hyphenation{ % some comment
//	 ...
//	}
func (dict *Dictionary) readExceptions(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "}") {
			return
		}
		var positions []int // we'll extract positions
		washyphen := false
		for _, char := range line {
			if char == '-' {
				positions = append(positions, 1) // possible break point
				washyphen = true
			} else if washyphen { // skip letter
				washyphen = false
			} else { // a letter without a '-'
				positions = append(positions, 0) // append 0
			}
		}
		//word := strings.Replace(line, "-", "", -1)
		word := strings.ReplaceAll(line, "-", "")
		dict.exceptions[word] = positions
		//fmt.Printf("exception '%s'\thas positions %v\n", line, positions)
	}
}

// HyphenationString returns word with discretionary hyphens inserted.
// Example:
//
//	"table" => "ta-ble".
func (dict *Dictionary) HyphenationString(word string) string {
	s := dict.Hyphenate(word)
	return strings.Join(s, "-")
}

// Hyphenate splits word at legal hyphenation positions.
//
// Example:
//
//	"table" => [ "ta", "ble" ].
func (dict *Dictionary) Hyphenate(word string) []string {
	const (
		leftMin  = 2
		rightMin = 2
	)
	if dict == nil || dict.patterns == nil || dict.patternsV == nil {
		return []string{word}
	}
	if positions, found := dict.exceptions[word]; found {
		return splitAtPositions(word, positions)
	}
	dottedword := "." + word + "."
	var positions = make([]int, 10) // the resulting hyphenation positions
	l := len(dottedword)
	for i := range l { // "word", "ord", "rd", "d"
		positions = mergePrefixPositions(dottedword[i:l], dict.patterns, dict.patternsV, i, positions)
	}
	positions = positions[1 : len(positions)-1]
	for i := 0; i < leftMin && i < len(positions); i++ {
		positions[i] = 0 // disallow breaks too close to the left edge
	}
	rightCutoff := len(word) - rightMin + 1 // indices >= cutoff leave fewer than rightMin chars
	rightCutoff = max(0, rightCutoff)
	// if rightCutoff < 0 {
	// 	rightCutoff = 0
	// }
	for i := rightCutoff; i < len(positions); i++ {
		positions[i] = 0 // disallow breaks too close to the right edge
	}
	return splitAtPositions(word, positions)
}

// mergePrefixPositions looks up all prefixes of a fragment and merges matching
// pattern weights into positions at absolute offset at.
func mergePrefixPositions(wordfragment string, tr *compacttrie.TinyHashTrie, store *patternStore, at int,
	positions []int) []int {
	//
	key, ok := encodeTrieKey(wordfragment)
	if !ok {
		return positions
	}
	it := tr.Iterator()
	for _, c := range key {
		patternID := it.Next(int8(c))
		if patternID == 0 {
			break
		}
		positions = store.MergeInto(patternID, at, positions)
	}
	return positions
}

// Helper: split a string at positions given by an integer slice.
func splitAtPositions(word string, positions []int) []string {
	var pp = make([]string, 0, len(word)/3)
	prev := 0                       // holds the last split index
	for i, pos := range positions { // check every position
		if pos > 0 && pos%2 != 0 { // if position is odd > 0
			pp = append(pp, word[prev:i]) // append syllable
			prev = i                      // remember last split index
		}
	}
	pp = append(pp, word[prev:]) // append last syllable
	return pp
}
