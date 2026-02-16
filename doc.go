/*
Package hyphenation is a quick and dirty implementation of a hyphenation algorithm.

(TODO: make more time and space efficient).

Package for an algorithm to hyphenate words. It is based on an algorithm
described by Frank Liang (F.M.Liang http://www.tug.org/docs/liang/). It loads
a pattern file (available with the TeX distribution) and builds a trie
structure, carrying an array of positions at every 'leaf'.

The trie package I'm using here is a very naive implementation and should
be replaced by a more sophisticated one
(e.g., https://github.com/siongui/go-succinct-data-structure-trie).
Resulting from the API of the trie, the implementation of the pattern
application algorithm is bad. TODO: improve time complexity of pattern
application.

Further Reading

  https://www.microsoft.com/en-us/Typography/OpenTypeSpecification.aspx
  https://nedbatchelder.com/code/modules/hyphenate.html   (Python implementation)
  http://www.mnn.ch/hyph/hyphenation2.html  / https://github.com/mnater/hyphenator

----------------------------------------------------------------------

BSD License

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
