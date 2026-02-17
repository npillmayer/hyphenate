package tex

import (
	"bytes"
	"io"

	"github.com/npillmayer/hyphenate"
	"github.com/npillmayer/hyphenate/tex/texexceptions"
	"github.com/npillmayer/hyphenate/tex/texpatterns"
)

// LoadDictionary loads a pattern dictionary and an exception list in TeX format.
//
// Please refer to
//
//	https://github.com/hyphenation/tex-hyphen/tree/master/hyph-utf8/tex/generic/hyph-utf8/patterns/tex
//
// for a list of real-workd pattern-files.
//
// Example usage:
//
//	f, _ := os.Open("path/to/patterns/hyph-en-us.tex")
//	defer f.Close()
//
//	dict := tex.LoadDictionary("en-us", f)
//
// This will load the patterns and exceptions temporarily into memory
func LoadDictionary(name string, reader io.Reader) (*hyphenate.Dictionary, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	texreader := texpatterns.NewPatternReader(bytes.NewReader(data))
	dict, err := hyphenate.LoadPatterns(name, texreader)
	if err != nil {
		return nil, err
	}
	err = dict.LoadExceptions(texexceptions.NewReader(bytes.NewReader(data)))
	return dict, err
}
