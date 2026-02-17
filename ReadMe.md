# hyphenate

`hyphenate` is a Go package for dictionary-based word hyphenation.
The base package compiles format-agnostic patterns and exceptions into a
compact DAT-backed lookup structure.

## Purpose

This package is intended for:

- compiling hyphenation patterns from streaming sources
- loading explicit hyphenation exceptions
- computing hyphenation opportunities in words
- keeping runtime memory compact for read-many workloads

## Base Package API (`github.com/npillmayer/hyphenate`)

### `type Pattern`

Format-agnostic pattern entry:

- `Sequence []rune`
- `Weights []int`

### `type PatternReader`

Streaming interface for pattern sources:

- `Next() (sequence []rune, weights []int, err error)`

### `type ExceptionReader`

Streaming interface for exception sources:

- `Next() (word string, positions []int, err error)`

### `func LoadPatterns(name string, reader PatternReader) (*Dictionary, error)`

Compiles patterns from a streaming reader into a dictionary.

### `func (dict *Dictionary) LoadExceptions(reader ExceptionReader) error`

Loads exceptions from a streaming reader.

### `func (dict *Dictionary) LoadExceptionList(exceptions map[string][]int)`

Loads exceptions from an in-memory map.

### `func (dict *Dictionary) AddException(word string, positions []int)`

Adds one explicit exception.

### `func (dict *Dictionary) Hyphenate(word string) []string`

Returns `word` split at legal hyphenation positions.

### `func (dict *Dictionary) HyphenationString(word string) string`

Returns `word` with `-` inserted at legal hyphenation positions.

## TeX Sub-Packages

TeX parsing is intentionally outside the base package:

- convenience API: `github.com/npillmayer/hyphenate/tex`
- patterns parser: `github.com/npillmayer/hyphenate/tex/texpatterns`
- exceptions parser: `github.com/npillmayer/hyphenate/tex/texexceptions`

Use these adapters when loading TeX `\patterns{...}` and `\hyphenation{...}`
files.

## Example: Base API (Format-Agnostic)

```go
package main

import (
	"fmt"
	"io"

	"github.com/npillmayer/hyphenate"
)

type emptyPatterns struct{}

func (emptyPatterns) Next() ([]rune, []int, error) {
	return nil, nil, io.EOF
}

func main() {
	dict, err := hyphenate.LoadPatterns("empty", emptyPatterns{})
	if err != nil {
		panic(err)
	}
	dict.AddException("table", []int{0, 0, 1, 0, 0})
	fmt.Println(dict.HyphenationString("table")) // ta-ble
}
```

## Example: TeX File Loading

```go
package main

import (
	"fmt"
	"os"

	"github.com/npillmayer/hyphenate/tex"
)

func main() {
	f, err := os.Open("testdata/hyph-en-us.tex")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dict, err := tex.LoadDictionary("hyph-en-us.tex", f)
	if err != nil {
		panic(err)
	}

	fmt.Println(dict.HyphenationString("algorithm")) // al-go-rithm
}
```

## Notes

- Pattern matching is Unicode-aware for BMP characters.
- Exceptions are applied before pattern-based hyphenation.
