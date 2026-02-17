/*
Package hyphenation is a quick and dirty implementation of a hyphenation algorithm.

(TODO: make more time and space efficient).

Package for an algorithm to hyphenate words. It is based on an algorithm
described by Frank Liang (F.M.Liang http://www.tug.org/docs/liang/). It loads
a pattern file (available with the TeX distribution) and builds a frozen
double-array trie (DAT) index. Hyphenation weight vectors are stored separately
in a compact payload store and referenced by trie state IDs.

The lookup path is Unicode-aware for BMP characters and supports non-ASCII
patterns such as German umlauts.

Further Reading

	https://www.microsoft.com/en-us/Typography/OpenTypeSpecification.aspx
	https://nedbatchelder.com/code/modules/hyphenate.html   (Python implementation)
	http://www.mnn.ch/hyph/hyphenation2.html  / https://github.com/mnater/hyphenator

----------------------------------------------------------------------

# BSD License

Copyright (c) Norbert Pillmayer <norbert@pillmayer@com>

All rights reserved.

License information is available in the LICENSE file.
*/
package hyphenate

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer writes to trace with key 'hyphenate'
func tracer() tracing.Trace {
	return tracing.Select("hyphenate")
}

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}
