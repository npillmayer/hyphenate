package hyphenate

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Pattern is a format-agnostic hyphenation pattern representation.
//
// Sequence is the rune sequence to match (for example: ".ab", "für").
// Weights stores Liang weights by relative position and may be longer than
// Sequence by one entry when a pattern has a trailing weight digit.
type Pattern struct {
	Sequence []rune
	Weights  []int
}

// PatternReader yields compiled pattern entries one-by-one.
// It should return io.EOF when the stream is exhausted.
type PatternReader interface {
	Next() (sequence []rune, weights []int, err error)
}

// ExceptionReader yields hyphenation exceptions one-by-one.
// It should return io.EOF when the stream is exhausted.
type ExceptionReader interface {
	Next() (word string, positions []int, err error)
}

// Dictionary is a loaded hyphenation dictionary.
//
// A dictionary contains:
//   - pattern rules (compiled into a pattern trie backend + compact metadata store)
//   - explicit hyphenation exceptions loaded through ExceptionReader.
type Dictionary struct {
	exceptions map[string][]int // e.g., "computer" => [3,5] = "com-pu-ter"
	patterns   patternTrie
	patternsV  *patternStore // compact metadata vectors by pattern id
	Identifier string        // Identifies the dictionary
}

// PatternTrieStats reports density metrics for the underlying pattern trie.
func (dict *Dictionary) PatternTrieStats() (backend string, usedSlots, totalSlots, maxStateID int, fillRatio float64) {
	if dict == nil || dict.patterns == nil {
		return "", 0, 0, 0, 0
	}
	stats := dict.patterns.Stats()
	return stats.Backend, stats.UsedSlots, stats.TotalSlots, stats.MaxStateID, stats.FillRatio()
}

// LoadPatterns compiles patterns from a streaming, format-agnostic source.
//
// File format parsing is intentionally outside the base package. Use adapters
// like package texpatterns to parse concrete formats and feed this API.
func LoadPatterns(name string, reader PatternReader) (dict *Dictionary, err error) {
	trie := mustNewDATBackend()
	type pendingPayload struct {
		pos    int
		packed []byte
	}
	pending := make([]pendingPayload, 0, 1024)
	maxPacked := 0
	dict = &Dictionary{
		exceptions: make(map[string][]int),
		patterns:   trie,
		Identifier: fmt.Sprintf("patterns: %s", name),
	}
	var sequence []rune
	var weights []int
	for {
		sequence, weights, err = reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}
		key, ok := dict.patterns.EncodeKey(string(sequence))
		if !ok {
			continue // simply skip invalid patterns
		}
		pos := dict.patterns.AllocPositionForWord(key)
		if pos == 0 {
			err = fmt.Errorf("could not allocate trie position for pattern %q", string(sequence))
			return
		}
		var packed []byte
		packed, err = packPositions(weights)
		if err != nil {
			return
		}
		if len(packed) > maxPacked {
			maxPacked = len(packed)
		}
		pending = append(pending, pendingPayload{pos: pos, packed: packed})
	}
	dict.patterns.Freeze()
	dict.patternsV = newPatternStore(uint8(maxPacked))
	for _, p := range pending {
		patternID := dict.patterns.ResolvePosition(p.pos)
		if patternID == 0 {
			err = fmt.Errorf("could not resolve trie position after freeze for temporary position %d", p.pos)
			return
		}
		if err = dict.patternsV.PutPacked(patternID, p.packed); err != nil {
			return
		}
	}
	backend, used, total, maxStateID, fill := dict.PatternTrieStats()
	tracer().Infof("pattern trie stats backend=%s used=%d total=%d fill=%.2f maxStateID=%d",
		backend, used, total, fill, maxStateID)
	return dict, nil
}

// LoadExceptions loads exception entries from a streaming source.
func (dict *Dictionary) LoadExceptions(reader ExceptionReader) (err error) {
	for {
		var word string
		var positions []int
		word, positions, err = reader.Next()
		if err == io.EOF {
			return nil
		} else if err != nil {
			break
		}
		dict.AddException(word, positions)
	}
	return err
}

// LoadExceptionList loads explicit exception entries from an in-memory map.
func (dict *Dictionary) LoadExceptionList(exceptions map[string][]int) {
	for word, positions := range exceptions {
		dict.AddException(word, positions)
	}
}

// AddException registers one explicit hyphenation exception.
func (dict *Dictionary) AddException(word string, positions []int) {
	if dict.exceptions == nil {
		dict.exceptions = make(map[string][]int)
	}
	pp := make([]int, len(positions))
	copy(pp, positions)
	dict.exceptions[word] = pp
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
	wordRunes := []rune(word)
	dottedword := make([]rune, 0, len(wordRunes)+2)
	dottedword = append(dottedword, '.')
	dottedword = append(dottedword, wordRunes...)
	dottedword = append(dottedword, '.')
	positions := make([]int, len(dottedword)) // the resulting hyphenation positions
	for i := range len(dottedword) {          // "word", "ord", "rd", "d"
		positions = mergePrefixPositions(string(dottedword[i:]), dict.patterns, dict.patternsV, i, positions)
	}
	positions = positions[1 : len(positions)-1]
	for i := 0; i < leftMin && i < len(positions); i++ {
		positions[i] = 0 // disallow breaks too close to the left edge
	}
	rightCutoff := len(wordRunes) - rightMin + 1 // indices >= cutoff leave fewer than rightMin chars
	rightCutoff = max(0, rightCutoff)
	for i := rightCutoff; i < len(positions); i++ {
		positions[i] = 0 // disallow breaks too close to the right edge
	}
	return splitAtPositions(word, positions)
}

// mergePrefixPositions looks up all prefixes of a fragment and merges matching
// pattern weights into positions at absolute offset at.
func mergePrefixPositions(wordfragment string, tr patternTrie, store *patternStore, at int,
	positions []int) []int {
	//
	key, ok := tr.EncodeKey(wordfragment)
	if !ok {
		return positions
	}
	it := tr.Iterator()
	for _, c := range key {
		patternID := it.Next(c)
		if patternID == 0 {
			break
		}
		positions = store.MergeInto(patternID, at, positions)
	}
	return positions
}

// Helper: split a string at positions given by an integer slice.
func splitAtPositions(word string, positions []int) []string {
	offsets := runeByteOffsets(word)
	runeCount := len(offsets) - 1
	var pp = make([]string, 0, max(1, runeCount/3))
	prev := 0                       // holds the last split index
	for i, pos := range positions { // check every position
		if i <= 0 || i >= runeCount {
			continue
		}
		if pos > 0 && pos%2 != 0 { // if position is odd > 0
			split := offsets[i]
			pp = append(pp, word[prev:split]) // append syllable
			prev = split                      // remember last split index
		}
	}
	pp = append(pp, word[prev:]) // append last syllable
	return pp
}

func runeByteOffsets(s string) []int {
	offsets := make([]int, 0, utf8.RuneCountInString(s)+1)
	for i := range s {
		offsets = append(offsets, i)
	}
	offsets = append(offsets, len(s))
	return offsets
}
