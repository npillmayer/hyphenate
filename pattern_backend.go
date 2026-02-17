package hyphenate

// patternIterator iterates over successive prefix states for one key.
type patternIterator interface {
	Next(symbol uint16) int
}

type patternTrieStats struct {
	Backend    string
	UsedSlots  int
	TotalSlots int
	MaxStateID int
}

func (s patternTrieStats) FillRatio() float64 {
	if s.TotalSlots == 0 {
		return 0
	}
	return float64(s.UsedSlots) / float64(s.TotalSlots)
}

// patternTrie is the internal backend abstraction for pattern-key storage.
type patternTrie interface {
	EncodeKey(s string) ([]uint16, bool)
	AllocPositionForWord(key []uint16) int
	ResolvePosition(pos int) int
	Freeze()
	Iterator() patternIterator
	Stats() patternTrieStats
}
