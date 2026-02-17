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

## Base Package API

A `hyphenate.Dictionary` lets you find hyphenation opportunities in words.

```go
	hyphenated := dictEN.HyphenationString("computer"))
	fmt.Println("computer => %s", hyphenated)
	segments := dictDE.Hyphenate("Fürsorge") // German
	fmt.Println("Fürsorge => %v", segments)
```

will result in

```shell
% computer => com-put-er
  Fürsorge => [ "Für", "sor", "ge" ].
```

- Pattern matching is Unicode-aware for BMP characters.
- Exceptions are applied before pattern-based hyphenation (see below).


### Loading Hyphenation Patterns

Before a dictionary can be used, it must be initialized with hyphenation patterns.
Clients load hyphenation patterns for each language they want to support. Every
dictionary is associated with exactly one language.
A dictionary loads hyphenation patterns from a source, usually a pattern-file.

```go
  (*Dictionary).LoadPatterns(name string, reader PatternSource) error
```

#### Type `Pattern`, `PatternSource`

`Pattern`s are format-agnostic hyphenation pattern entries.
A `PatternReader` is a streaming interface for pattern sources.

### Loading Hyphenation Exceptions

Patterns will ususally only get you so far. Exceptions are needed to handle
special cases like "ta-ble" or "co-op-eration".

#### Type `ExceptionsSource`

An exception-source is a streaming interface for reading hyphenation-exceptions.

```go
  (*Dictionary).LoadExceptions(reader ExceptionReader) error
```

## TeX Sub-Packages

The TeX communitiy provides pattern files for a lot of languages
[on GitHub](https://github.com/hyphenation/tex-hyphen/tree/master/hyph-utf8/tex/generic/hyph-utf8/patt).
Please check the license before using them (most should be compatible with the MIT license).

- convenience API: `github.com/npillmayer/hyphenate/tex`
- patterns parser: `github.com/npillmayer/hyphenate/tex/texpatterns`
- exceptions parser: `github.com/npillmayer/hyphenate/tex/texexceptions`

Use these adapters when loading TeX `\patterns{...}` and `\hyphenation{...}`
files.

## Example: TeX Pattern-File Loading

```go
import "github.com/npillmayer/hyphenate/tex"

// hyph-en-us.tex contains patterns and exceptions for US-English
f, _ := os.Open("/path/to/patterns/hyph-en-us.tex")
defer f.Close()

dictEN, err := tex.LoadDictionary("en-US", f)
if err != nil {
	panic(err)
}

fmt.Println(dictEN.HyphenationString("algorithm")) // al-go-rithm
```
