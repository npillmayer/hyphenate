# dat

`dat` contains low-level data structures for a frozen double-array trie (DAT)
and BMP-aware rune mapping.

Import path:

- `github.com/npillmayer/hyphenate/dat`

## Scope

- `DAT`: compact base/check transition arrays plus terminal payload storage.
- `PagedMapBMP`: paged mapping from BMP code units to dense alphabet IDs.

This package is primarily an implementation detail used by the root
`hyphenate` package. Most users should use `github.com/npillmayer/hyphenate`
and, for TeX files, `github.com/npillmayer/hyphenate/tex`.

