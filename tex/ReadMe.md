# tex

`tex` provides convenience loading for TeX hyphenation dictionaries.

Import path:

- `github.com/npillmayer/hyphenate/tex`

## API

- `func LoadDictionary(name string, reader io.Reader) (*hyphenate.Dictionary, error)`

Loads both TeX patterns (`\patterns{...}`) and TeX exceptions
(`\hyphenation{...}`) from one source.

## Related TeX Sub-Packages

- patterns-only parser: `github.com/npillmayer/hyphenate/tex/texpatterns`
- exceptions-only parser: `github.com/npillmayer/hyphenate/tex/texexceptions`

Use those sub-packages directly when you need separate control over pattern and
exception loading phases.
