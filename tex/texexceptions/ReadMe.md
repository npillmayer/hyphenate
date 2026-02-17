# texexceptions

`texexceptions` parses TeX hyphenation exception blocks (`\hyphenation{...}`)
and loads them into an existing dictionary.

Import path:

- `github.com/npillmayer/hyphenate/tex/texexceptions`

## API

- `func LoadExceptions(dict *hyphenate.Dictionary, reader io.Reader)`

Parses exceptions from TeX input and adds them to `dict`.

- `func NewReader(reader io.Reader) *Reader`

Creates a streaming parser implementing the base package `ExceptionReader`
interface.

## Related TeX Packages

- patterns parser companion:
  `github.com/npillmayer/hyphenate/tex/texpatterns`
- convenience one-shot loader:
  `github.com/npillmayer/hyphenate/tex` (`LoadDictionary`)

Typical split usage:

1. load patterns with `texpatterns.LoadPatterns`
2. then load exceptions with `texexceptions.LoadExceptions`
