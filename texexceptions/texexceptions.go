package texexceptions

import (
	"bufio"
	"io"
	"strings"

	"github.com/npillmayer/hyphenate"
)

// Reader streams hyphenation exceptions from TeX \hyphenation{...} blocks.
type Reader struct {
	scanner *bufio.Scanner
	inBlock bool
}

// LoadExceptions parses TeX exception data from reader and adds all
// \hyphenation{...} entries to this dictionary.
func LoadExceptions(dict *hyphenate.Dictionary, reader io.Reader) {
	dict.LoadExceptionReader(NewReader(reader))
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{
		scanner: bufio.NewScanner(reader),
	}
}

// Next returns the next exception as (word, positions).
// It returns io.EOF when exhausted.
func (r *Reader) Next() (string, []int, error) {
	for r.scanner.Scan() {
		line := r.scanner.Text()
		if !r.inBlock {
			if strings.HasPrefix(line, "\\hyphenation{") {
				r.inBlock = true
			}
			continue
		}
		if strings.HasPrefix(line, "}") {
			r.inBlock = false
			continue
		}
		positions := make([]int, 0, len(line))
		wasHyphen := false
		for _, ch := range line {
			if ch == '-' {
				positions = append(positions, 1)
				wasHyphen = true
			} else if wasHyphen {
				wasHyphen = false
			} else {
				positions = append(positions, 0)
			}
		}
		word := strings.ReplaceAll(line, "-", "")
		return word, positions, nil
	}
	if err := r.scanner.Err(); err != nil {
		return "", nil, err
	}
	return "", nil, io.EOF
}
