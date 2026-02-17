package hyphenate

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/npillmayer/hyphenate/dat"
)

type datBuildNode struct {
	tmpID    int
	state    uint32
	children map[uint16]*datBuildNode
}

type datBackend struct {
	frozen      bool
	root        *datBuildNode
	nextNodeID  int
	runeToDense map[rune]uint16
	nextDenseID uint16
	compiled    *dat.DAT
}

func newDATBackend() *datBackend {
	backend := &datBackend{
		root:        &datBuildNode{tmpID: 1, children: make(map[uint16]*datBuildNode)},
		nextNodeID:  2,
		runeToDense: make(map[rune]uint16),
		nextDenseID: 1, // reserve 1 for '.'
		compiled: &dat.DAT{
			Root: 1,
		},
	}
	backend.runeToDense['.'] = 1
	backend.compiled.MapPaged.Set(uint16('.'), 1)
	return backend
}

func mustNewDATBackend() patternTrie {
	return newDATBackend()
}

func (db *datBackend) EncodeKey(s string) ([]uint16, bool) {
	key := make([]uint16, 0, utf8.RuneCountInString(s))
	if db.frozen {
		for _, r := range s {
			if r > 0xFFFF {
				key = append(key, 0)
				continue
			}
			key = append(key, db.compiled.Dense(uint16(r)))
		}
		return key, true
	}
	for _, r := range s {
		if r > 0xFFFF {
			return nil, false
		}
		dense, ok := db.runeToDense[r]
		if !ok {
			if db.nextDenseID == ^uint16(0) {
				return nil, false
			}
			db.nextDenseID++
			dense = db.nextDenseID
			db.runeToDense[r] = dense
			db.compiled.MapPaged.Set(uint16(r), dense)
		}
		key = append(key, dense)
	}
	return key, true
}

func (db *datBackend) AllocPositionForWord(key []uint16) int {
	if len(key) == 0 {
		return 0
	}
	if !db.frozen {
		n := db.root
		for _, c := range key {
			if c == 0 {
				return 0
			}
			child := n.children[c]
			if child == nil {
				child = &datBuildNode{
					tmpID:    db.nextNodeID,
					children: make(map[uint16]*datBuildNode),
				}
				db.nextNodeID++
				n.children[c] = child
			}
			n = child
		}
		return n.tmpID
	}
	state := db.compiled.Root
	for _, c := range key {
		if c == 0 {
			return 0
		}
		next, ok := db.compiled.Transition(state, c)
		if !ok {
			return 0
		}
		state = next
	}
	return int(state)
}

func (db *datBackend) Freeze() {
	if db.frozen {
		return
	}
	db.compiled.Sigma = db.nextDenseID
	db.compiled.Base = make([]int32, int(db.compiled.Root)+1)
	db.compiled.Check = make([]int32, int(db.compiled.Root)+1)
	db.root.state = db.compiled.Root
	queue := []*datBuildNode{db.root}
	for q := 0; q < len(queue); q++ {
		n := queue[q]
		if len(n.children) == 0 {
			continue
		}
		labels := sortedLabels(n.children)
		base := findDATBase(db.compiled.Check, labels)
		ensureDATIndex(db.compiled, base+int(labels[len(labels)-1]))
		db.compiled.Base[n.state] = int32(base)
		for _, label := range labels {
			t := base + int(label)
			ensureDATIndex(db.compiled, t)
			child := n.children[label]
			child.state = uint32(t)
			db.compiled.Check[t] = int32(n.state)
			queue = append(queue, child)
		}
	}
	db.compiled.PayloadOff = make([]uint32, len(db.compiled.Base))
	db.root = nil
	db.runeToDense = nil
	db.frozen = true
}

func (db *datBackend) Iterator() patternIterator {
	if db.frozen {
		return &datIterator{
			d:     db.compiled,
			state: db.compiled.Root,
		}
	}
	return &datBuildIterator{
		node: db.root,
	}
}

type datBuildIterator struct {
	node *datBuildNode
	dead bool
}

func (it *datBuildIterator) Next(symbol uint16) int {
	if it.dead || it.node == nil || symbol == 0 {
		it.dead = true
		return 0
	}
	next := it.node.children[symbol]
	if next == nil {
		it.dead = true
		return 0
	}
	it.node = next
	return next.tmpID
}

type datIterator struct {
	d     *dat.DAT
	state uint32
	dead  bool
}

func (it *datIterator) Next(symbol uint16) int {
	if it.dead || it.d == nil || symbol == 0 {
		it.dead = true
		return 0
	}
	next, ok := it.d.Transition(it.state, symbol)
	if !ok {
		it.dead = true
		return 0
	}
	it.state = next
	return int(next)
}

func sortedLabels(children map[uint16]*datBuildNode) []uint16 {
	labels := make([]uint16, 0, len(children))
	for label := range children {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool {
		return labels[i] < labels[j]
	})
	return labels
}

func findDATBase(check []int32, labels []uint16) int {
	for base := 1; ; base++ {
		ok := true
		for _, label := range labels {
			t := base + int(label)
			if t < len(check) && check[t] != 0 {
				ok = false
				break
			}
		}
		if ok {
			return base
		}
	}
}

func ensureDATIndex(d *dat.DAT, idx int) {
	if idx < len(d.Base) {
		return
	}
	grow := idx + 1 - len(d.Base)
	d.Base = append(d.Base, make([]int32, grow)...)
	d.Check = append(d.Check, make([]int32, grow)...)
	if len(d.PayloadOff) > 0 {
		d.PayloadOff = append(d.PayloadOff, make([]uint32, grow)...)
	}
}

func (db *datBackend) String() string {
	return fmt.Sprintf("DAT(states=%d,sigma=%d,frozen=%v)", db.compiled.NStates(), db.compiled.Sigma, db.frozen)
}

func (db *datBackend) Stats() patternTrieStats {
	stats := patternTrieStats{
		Backend:    "dat",
		TotalSlots: db.compiled.NStates(),
		MaxStateID: int(db.compiled.Root),
	}
	if stats.TotalSlots == 0 {
		return stats
	}
	used := 0
	maxID := int(db.compiled.Root)
	for i := range db.compiled.Check {
		if i == int(db.compiled.Root) || db.compiled.Check[i] != 0 {
			used++
			if i > maxID {
				maxID = i
			}
		}
	}
	stats.UsedSlots = used
	stats.MaxStateID = maxID
	return stats
}
