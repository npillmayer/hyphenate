package hyphenate

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/derekparker/trie" // TODO: replace this
)


// Dictionary is a type representing a hyphenation dictionary.
// A hyphenation dictionary consists of hyphenation patterns and a list of exceptions

type Dictionary struct {
	exceptions map[string][]int // e.g., "computer" => [3,5] = "com-pu-ter"
	patterns   *trie.Trie       // where we store patterns and positions
	Identifier string           // Identifies the dictionary
}

// LoadPatterns loads a pattern file. Returns the identifier of the pattern file and a trie.
//
// Patterns are enclosed in between
//
//    \patterns{ % some comment
//     ...
//    .wil5i
//    .ye4
//    4ab.
//    a5bal
//    a5ban
//    abe2
//     ...
//    }
//
// Odd numbers stand for possible discretionnaries, even numbers forbid
// hyphenation. Digits belong to the character immediately after them, i.e.,
//
//    "a5ban" => (a)(5b)(a)(n) => positions["aban"] = [0,5,0,0].
//
func LoadPatterns(name string, reader io.Reader) *Dictionary {
	dict := &Dictionary{
		exceptions: make(map[string][]int),
		patterns:   trie.New(),
		Identifier: fmt.Sprintf("patterns: %s", name),
	}
	scanner := bufio.NewScanner(reader)
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
			var pattern string          // will become the pattern without positions
			var positions []int         // we'll extract positions
			var wasdigit bool           // has the last char been a digit?
			for _, char := range line { // iterate over runes of the pattern
				if unicode.IsDigit(char) {
					d, _ := strconv.Atoi(string(char))
					positions = append(positions, d) // add to positions array
					wasdigit = true
				} else { // '.' or alphabetic rune
					pattern = pattern + string(char)
					if wasdigit {
						wasdigit = false
					} else {
						positions = append(positions, 0) // append a 0
					}
				}
			}
			//fmt.Printf("pattern '%s'\thas positions %v\n", pattern, positions)
			dict.patterns.Add(pattern, positions)
		}
	}
	return dict
}

// readExceptions reads exceptions from a pattern file. Exceptions are denoted as
//
//    ex-cep-tion
//    ta-ble
//
// and so on, a single word per line. Exceptions are enclosed in
//
//    \hyphenation{ % some comment
//     ...
//    }
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
		word := strings.Replace(line, "-", "", -1)
		dict.exceptions[word] = positions
		//fmt.Printf("exception '%s'\thas positions %v\n", line, positions)
	}
}

/*

// HyphenationString return a word with possible hyphens inserted.
// Example:
//
//   "table" => "ta-ble".
//
func (dict *Dictionary) HyphenationString(word string) string {
	s := dict.Hyphenate(word)
	return strings.Join(s, "-")
}

// Hyphenate returns a word split up at legal hyphenation positions.
//
// Example:
//
//     "table" => [ "ta", "ble" ].
//
func (dict *Dictionary) Hyphenate(word string) []string {
	if positions, found := dict.exceptions[word]; found {
		return splitAtPositions(word, positions)
	}
	dottedword := "." + word + "."
	var positions = make([]int, 10) // the resulting hyphenation positions
	l := len(dottedword)
	for i := 0; i < l; i++ { // "word", "ord", "rd", "d"
		pp := findPrefixPositions(dottedword[i:l], dict.patterns)
		positions = mergePositions(positions, pp, i)
	}
	positions = positions[1 : len(positions)-1]
	if positions[0] > 0 { // sometimes hyphen before first letter is "allowed"
		positions[0] = 0
	} else if len(positions) > len(word) && positions[len(word)] > 0 {
		positions[len(word)] = 0 // sometimes hyphen after last letter is "allowed"
	}
	return splitAtPositions(word, positions)
}

// For a given word fragment w, find trie entries for every prefix of w.
// Example: word fragement is "gment". Find entries for "gment", "ment",
// "ent", "nt", "t". Collect and return the resulting positions arrays.
func findPrefixPositions(wordfragment string, trie *trie.Trie) [][]int {
	l := len(wordfragment)
	var pp [][]int           // return value
	for j := 1; j < l; j++ { // for every char position until end of fragment
		node, _ := trie.Find(wordfragment[:j])
		if node != nil { // yes, entry found
			pp = append(pp, node.Meta().([]int)) // Meta() is arr of positions
		}
	}
	return pp
}

/*
Merge a collection of positions arrays to a given positions array at
a given index. Positions are overwritten, if a new position is greater
than the old one. If the positions array isn't long enough, it will be
enlarged.

Example:

    given p = [0,2,0,0] and pp = { [1,7], [0,0,3] } =>

after merge at position 1:

    p = [0,2,7,3].

*/
func mergePositions(positions []int, pp [][]int, at int) []int {
	for _, p := range pp { // for all partial position arrays
		for relAt, num := range p { // for evey relative position
			if at+relAt >= len(positions) { // array long enough?
				positions = append(positions, 0)
			}
			if num > positions[at+relAt] { // new pos greater than current pos?
				positions[at+relAt] = num
			}
		}
	}
	return positions
}

/* Helper: split a string at positions given by an integer slice.
 */
func splitAtPositions(word string, positions []int) []string {
	var pp = make([]string, 0, len(word)/3)
	prev := 0                       // holds the last split index
	for i, pos := range positions { // check every position
		if pos > 0 && pos%2 != 0 { // if position is odd > 0
			pp = append(pp, word[prev:i]) // append syllable
			prev = i                      // remember last split index
		}
	}
	pp = append(pp, word[prev:len(word)]) // append last syllable
	return pp
}

/* Helper: Insert hyphens in a string at positions given by an integer slice.
 */
func insertHyphens(word string, positions []int) string {
	hyphencount := 0
	for i, pos := range positions {
		if pos > 0 && pos%2 != 0 { // only odd positions are valid breakpoints
			//word = xstrings.Insert(word, "-", hyphencount+i)
			word = insertStr(word, "-", hyphencount+i)
			hyphencount++ // remember hyphens => hyphens prolong word
		}
	}
	return word
}

func insertStr(word string, s string, at int) string {
	return word[:at] + s + word[at:]
}
