package texpatterns

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/npillmayer/hyphenate"
)

// PatternReader streams Liang patterns from TeX-style source files.
type PatternReader struct {
	scanner    *bufio.Scanner
	identifier string
	sequence   []rune
	weights    []int
}

// LoadPatterns parses TeX pattern data and returns a ready-to-use dictionary.
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
// The loader parses TeX input into a streaming PatternReader and compiles
// patterns incrementally.
//
// Exceptions from \hyphenation{...} are intentionally not loaded here.
func LoadPatterns(name string, reader io.Reader) (*hyphenate.Dictionary, error) {
	r := NewPatternReader(reader)
	return hyphenate.LoadPatterns(name, r)
}

func NewPatternReader(reader io.Reader) *PatternReader {
	return &PatternReader{
		scanner:  bufio.NewScanner(reader),
		sequence: make([]rune, 0, 32),
		weights:  make([]int, 0, 32),
	}
}

func (r *PatternReader) Identifier() string {
	return r.identifier
}

// Next returns the next pattern as (sequence, weights).
// It returns io.EOF when exhausted.
// The returned slices are reused by subsequent calls.
func (r *PatternReader) Next() ([]rune, []int, error) {
	var err error
	inPatterns := false
	for r.scanner.Scan() {
		line := r.scanner.Text()
		if strings.HasPrefix(line, "%     message: ") {
			r.identifier = line[15:]
			continue
		}
		if strings.HasPrefix(line, "\\message{") {
			r.identifier = line[9 : len(line)-1]
			continue
		}
		if strings.HasPrefix(line, "\\hyphenation{") {
			skipTeXBlock(r.scanner)
			continue
		}
		if strings.HasPrefix(line, "%") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "\\patterns{") {
			inPatterns = true
			continue
		}
		if strings.HasPrefix(line, "}") {
			if inPatterns { // closing of patterns block
				return nil, nil, io.EOF // do not read further
			}
			continue
		}
		r.decodePatternLine(line)
		if len(r.sequence) == 0 {
			continue
		}
		return r.sequence, r.weights, nil
	}
	if err = r.scanner.Err(); err != nil {
		return nil, nil, err
	}
	if inPatterns {
		err = errors.New("unexpected end of file (unclosed \\patterns block)")
	}
	return nil, nil, io.EOF
}

func (r *PatternReader) decodePatternLine(line string) {
	r.sequence = r.sequence[:0]
	r.weights = r.weights[:0]
	wasDigit := false
	for _, ch := range line {
		if unicode.IsDigit(ch) {
			d, _ := strconv.Atoi(string(ch))
			r.weights = append(r.weights, d)
			wasDigit = true
			continue
		}
		r.sequence = append(r.sequence, ch)
		if wasDigit {
			wasDigit = false
		} else {
			r.weights = append(r.weights, 0)
		}
	}
}

func skipTeXBlock(scanner *bufio.Scanner) {
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "}") {
			return
		}
	}
}
