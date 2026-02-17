# hyphenate

`hyphenate` is a Go package for dictionary-based word hyphenation.
It reads TeX hyphenation pattern files, compiles them into a compact DAT-backed
lookup structure, and returns legal hyphenation split points for words.

## Purpose

This package is intended for:

- loading TeX hyphenation pattern dictionaries (`\patterns{...}` and `\hyphenation{...}`)
- computing hyphenation opportunities in words
- doing this with compact memory usage for read-many workloads

Internally, the package compiles pattern keys into a frozen double-array trie (DAT)
with BMP-aware symbol mapping and stores pattern weights in a packed metadata store.

## Package API

### `func LoadPatterns(name string, reader io.Reader) *Dictionary`

Parses and compiles a hyphenation dictionary from `reader`.

- `name` is used for dictionary identification.
- Returns a `*Dictionary` ready for lookup.
- Exceptions from `\hyphenation{...}` are loaded and take precedence.

### `type Dictionary`

Loaded dictionary handle for hyphenation lookups.

Exported field:

- `Identifier string`: dictionary identifier extracted from input (or fallback name)

### `func (dict *Dictionary) Hyphenate(word string) []string`

Returns `word` split at legal hyphenation positions.

Example output: `["al", "go", "rithm"]`.

### `func (dict *Dictionary) HyphenationString(word string) string`

Returns `word` with `-` inserted at legal hyphenation positions.

Example output: `"al-go-rithm"`.

## Example

```go
package main

import (
	"fmt"
	"os"

	"github.com/npillmayer/hyphenate"
)

func main() {
	f, err := os.Open("testdata/hyph-en-us.tex")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dict := hyphenate.LoadPatterns("hyph-en-us.tex", f)

	fmt.Println(dict.HyphenationString("algorithm")) // al-go-rithm
	fmt.Println(dict.Hyphenate("concatenation"))     // [con cate na tion]
}
```

## Notes

- Pattern matching is Unicode-aware for BMP characters (e.g., `ä`, `ö`, `ü`, `ß`).
- Exceptions are applied before pattern-based hyphenation.
