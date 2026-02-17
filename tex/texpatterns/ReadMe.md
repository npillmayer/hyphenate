# texpatterns

`texpatterns` parses TeX/Liang hyphenation patterns and compiles them into a
`hyphenate.Dictionary`.

Import path:

- `github.com/npillmayer/hyphenate/tex/texpatterns`

## API

### `func LoadPatterns(name string, reader io.Reader) (*hyphenate.Dictionary, error)`

Parses `\patterns{...}` data and builds a dictionary from patterns.
It does not load TeX exceptions.

### `func NewPatternReader(reader io.Reader) *PatternReader`

Creates a streaming parser implementing the base package `PatternReader`
interface.

## Related TeX Packages

- exceptions parser companion:
  `github.com/npillmayer/hyphenate/tex/texexceptions`
- convenience one-shot loader:
  `github.com/npillmayer/hyphenate/tex` (`LoadDictionary`)

Typical split usage:

1. load patterns with `texpatterns.LoadPatterns`
2. then load exceptions with `texexceptions.LoadExceptions`

