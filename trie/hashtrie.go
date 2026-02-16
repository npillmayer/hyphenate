package trie

import (
	"errors"
	"math"
	"unsafe"
)

// type pointer int8
type pointer uint16
type category int8

const empty = 0         // denotes empty slots of the trie
const tolerance = 20000 // maximum number of tries to (re)locate a family

// TinyHashTrie is a compact, write-optimized hash-trie for categorical byte
// sequences. It is intended for write-once/read-many workloads.
type TinyHashTrie struct {
	frozen     bool
	headercode category
	span       pointer // space in table without leading and trailing #catnct slots
	alpha      pointer
	size       int
	catcnt     int
	link       []pointer
	sibling    []pointer
	ch         []category
}

// NewTinyHashTrie creates a new trie.
//
// size should be prime. catcnt is the maximum category value expected in input
// bytes (excluding 0, which is reserved internally).
func NewTinyHashTrie(size uint16, catcnt int8) (*TinyHashTrie, error) {
	//func NewTinyHashTrie(size int16, catcnt int8) (*TinyHashTrie, error) {
	if catcnt > 85 {
		tracer().Errorf("number of categories to store may not exceed 85")
		return nil, errors.New("number of categories to store may not exceed 85")
	}
	trie := &TinyHashTrie{
		size:   int(size),       // TODO find nearest prime
		catcnt: int(catcnt) + 1, // make room for cat = 0
	}
	tracer().Infof("hash trie size = %d for %d categories", trie.size, trie.catcnt-1)
	trie.headercode = category(trie.catcnt) + 1
	trie.alpha = pointer(math.Round(0.61803 * float64(trie.size)))
	trie.span = pointer(trie.size - 2*trie.catcnt)
	trie.setInitialValues()
	return trie, nil
}

func (trie *TinyHashTrie) setInitialValues() {
	trie.link = make([]pointer, trie.size)
	trie.sibling = make([]pointer, trie.size)
	trie.ch = make([]category, trie.size)
	for i := 1; i <= int(trie.catcnt); i++ {
		trie.ch[i] = category(i)
		trie.sibling[i] = pointer(i) - 1
	}
	trie.ch[0] = trie.headercode
	trie.sibling[0] = pointer(trie.catcnt)
}

// AllocPositionForWord inserts/looks up buf and returns its trie position.
//
// Behavior depends on mode:
//   - mutable (default): missing nodes are created
//   - frozen: lookup-only; returns 0 if buf is not present
func (trie *TinyHashTrie) AllocPositionForWord(buf []byte) int {
	n := 0
	p := pointer(trie.correct(buf[0])) // current word position
	for n+1 < len(buf) {
		n++
		c := trie.correct(buf[n])
		p = trie.advanceToChild(p, c, n)
		//T().Debugf("advanced to p=%d with c=%d", p, c)
	}
	return int(p)
}

func (trie *TinyHashTrie) advanceToChild(p pointer, c category, n int) pointer {
	if trie.link[p] == 0 {
		if trie.frozen {
			return 0
		}
		tracer().Debugf("link[%d] is unassigned, inserting first child=%d", p, c)
		return trie.insertFirstbornChildAndProgress(p, c, n)
	}
	//T().Debugf("position link[%d] is occupied →%d", p, trie.link[p])
	q := trie.link[p] + pointer(c)
	if trie.ch[q] != c {
		if trie.frozen {
			return 0
		}
		if trie.ch[q] != empty {
			p, q = trie.moveFamily(p, c, n)
		}
		q = trie.insertChildIntoFamily(p, q, c)
	}
	return q
}

func (trie *TinyHashTrie) insertFirstbornChildAndProgress(p pointer, c category, n int) pointer {
	var h pointer                               // trial header location
	var x = pointer(n) * trie.alpha % trie.span // nominal position of header #n
	h = x + pointer(trie.catcnt) + 1            // now catcnt < h ≤ (trie.size+catcnt)
	if h <= pointer(trie.catcnt) || int(h) > trie.size+trie.catcnt {
		panic("invariant not held")
	}
	trys := 0
	for ; trys < tolerance; trys++ {
		// Compute the next trial header location or abort find
		if h == pointer(trie.size-trie.catcnt) {
			h = pointer(trie.catcnt + 1)
		} else {
			h++
		}
		//
		if trie.ch[h] == empty && trie.ch[h+pointer(c)] == empty {
			tracer().Debugf("found an empty child slot=%d→%d", h, h+(pointer(c)))
			break
		}
	}
	if trys == tolerance {
		tracer().Errorf("abort find")
		panic("abort find")
	}
	trie.link[p], trie.link[h] = h, p
	p = h + pointer(c)
	trie.ch[h], trie.ch[p] = trie.headercode, c
	trie.sibling[h], trie.sibling[p] = p, h
	trie.link[p] = 0
	return p
}

// q = link[p] + c
func (trie *TinyHashTrie) insertChildIntoFamily(p, q pointer, c category) pointer {
	h := trie.link[p]
	for trie.sibling[h] > q {
		h = trie.sibling[h]
	}
	trie.sibling[q], trie.sibling[h] = trie.sibling[h], q
	trie.ch[q] = c
	trie.link[q] = 0
	tracer().Debugf("Inserted %d at q=%d with header=%d", c, q, trie.link[p])
	return q
}

func (trie *TinyHashTrie) moveFamily(p pointer, c category, n int) (pointer, pointer) {
	tracer().Debugf("have to move family for c=%d", c)
	//
	var h pointer                               // trial header location
	var x = pointer(n) * trie.alpha % trie.span // nominal position of header #n
	h = x + pointer(trie.catcnt) + 1            // now catcnt < h ≤ (trie.size+catcnt)
	if h <= pointer(trie.catcnt) || int(h) > trie.size+trie.catcnt {
		panic("invariant not held")
	}
	h = x + pointer(trie.catcnt) + 1
	//
	q := h + pointer(c)
	r := trie.link[p]
	delta := h - r
	trys := 0
	for ; trys < tolerance; trys++ {
		// Compute the next trial header location
		tracer().Debugf("trying to find a home for family")
		if h < pointer(trie.size-trie.catcnt) {
			h++
		} else {
			h = pointer(trie.catcnt) + 1
		}
		//
		if trie.ch[h+pointer(c)] != empty {
			continue
		}
		tracer().Debugf("found a potential home h=%d", h)
		r = trie.link[p]
		delta = h - r
		for trie.ch[r+delta] == empty && trie.sibling[r] != trie.link[p] {
			r = trie.sibling[r]
			tracer().Debugf(".")
		}
		if trie.ch[r+delta] == empty {
			break // found a slot
		}
	}
	if trys >= tolerance {
		tracer().Errorf("abort find")
		panic("abort find")
	}
	q = h + pointer(c)
	r = trie.link[p]
	delta = h - r
	for {
		trie.sibling[r+delta] = trie.sibling[r] + delta
		trie.ch[r+delta] = trie.ch[r]
		trie.ch[r] = empty
		trie.link[r+delta] = trie.link[r]
		if trie.link[r] != empty {
			trie.link[trie.link[r]] = r + delta
		}
		r = trie.sibling[r]
		if trie.ch[r] == empty {
			break
		}
	}
	return p, q
}

// Freeze makes the trie read-only and drops construction-only state.
func (trie *TinyHashTrie) Freeze() {
	trie.frozen = true
	trie.sibling = nil // will not be needed for lookup
}

// Iterator returns a stateful prefix iterator for lookup in this trie.
func (trie *TinyHashTrie) Iterator() *Iterator {
	iterator := &Iterator{
		trie: trie,
		n:    0,
	}
	return iterator
}

func (trie *TinyHashTrie) correct(c byte) category {
	if c == 0 { // bidi.L = 0 ⇒ unusable, set to max_cat
		return category(trie.catcnt - 1)
	}
	return category(c)
}

// --- Iterator --------------------------------------------------------------

// Iterator advances through successive prefix states for one query word.
type Iterator struct {
	trie     *TinyHashTrie
	position pointer
	n        int
}

// Next advances the iterator by one category byte and returns trie position.
// It returns 0 when the prefix is not present.
func (ti *Iterator) Next(c int8) int {
	if ti.trie == nil {
		return 0
	}
	if ti.n == 0 {
		ti.position = pointer(ti.trie.correct(byte(c)))
		ti.n++
		return int(ti.trie.ch[ti.position])
	}
	cc := ti.trie.correct(byte(c))
	ti.position = ti.trie.advanceToChild(ti.position, cc, ti.n)
	//T().Debugf("advanced to p=%d with c=%d", ti.position, c)
	if ti.position == 0 {
		ti.trie = nil // end of iteration
	}
	return int(ti.position)
}

// ---------------------------------------------------------------------------

// Stats writes capacity and memory statistics to the trace log.
func (trie *TinyHashTrie) Stats() {
	fillch := 0
	filllink := 0
	for i := 0; i < trie.size; i++ {
		if trie.ch[i] != empty {
			fillch++
		}
		if trie.link[i] != empty {
			filllink++
		}
	}
	tracer().Infof("Trie Statistics:")
	tracer().Infof("  Size of trie:   %d", trie.size)
	tracer().Infof("  Category count: %d", trie.catcnt)
	tracer().Infof("  Links:    %d of %d (%.1f%%)", filllink, trie.size, float32(filllink)/float32(trie.size)*100)
	tracer().Infof("  Children: %d of %d (%.1f%%)", fillch, trie.size, float32(fillch)/float32(trie.size)*100)
	var memory uint64
	memory = uint64(unsafe.Sizeof(*trie))
	test := pointer(1)
	word := int(unsafe.Sizeof(test))
	memory += uint64(trie.size * 2 * word)
	tracer().Infof("  Memory:   %d bytes", memory)
}
