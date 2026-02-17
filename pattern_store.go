package hyphenate

import "fmt"

const absentPayload = 0xFF
const initialPatternStoreSlots = 2 // include slot 0 + root slot

// patternStore keeps packed hyphenation vectors directly indexed by trie position.
// Each non-zero vector entry is stored as one byte: high nibble=index, low nibble=value.
type patternStore struct {
	width   uint8
	length  []uint8 // will grow with demand
	payload []byte  // will grow with demand
}

func packPositions(positions []int) ([]byte, error) {
	packed := make([]byte, 0, len(positions))
	for rel, val := range positions {
		if val == 0 {
			continue
		}
		if rel > 15 {
			return nil, fmt.Errorf("relative index out of range (0..15): %d", rel)
		}
		if val < 0 || val > 15 {
			return nil, fmt.Errorf("value out of range (0..15): %d", val)
		}
		packed = append(packed, byte((rel<<4)|val))
	}
	return packed, nil
}

func newPatternStore(maxPackedEntries uint8) *patternStore {
	if maxPackedEntries > 16 {
		maxPackedEntries = 16
	}
	s := &patternStore{
		width:   maxPackedEntries,
		length:  make([]uint8, initialPatternStoreSlots),
		payload: make([]byte, initialPatternStoreSlots*int(maxPackedEntries)),
	}
	for i := range s.length {
		s.length[i] = absentPayload
	}
	return s
}

func (s *patternStore) ensure(pos int) {
	if pos < len(s.length) {
		return
	}
	grow := pos + 1 - len(s.length)
	old := len(s.length)
	s.length = append(s.length, make([]uint8, grow)...)
	for i := old; i < len(s.length); i++ {
		s.length[i] = absentPayload
	}
	if s.width > 0 {
		s.payload = append(s.payload, make([]byte, grow*int(s.width))...)
	}
}

// Put stores a positions vector at trie position pos.
func (s *patternStore) Put(pos int, positions []int) error {
	if pos < 0 {
		return fmt.Errorf("negative trie position: %d", pos)
	}
	packed, err := packPositions(positions)
	if err != nil {
		return err
	}
	return s.PutPacked(pos, packed)
}

// PutPacked stores already-packed payload at trie position pos.
func (s *patternStore) PutPacked(pos int, packed []byte) error {
	if pos < 0 {
		return fmt.Errorf("negative trie position: %d", pos)
	}
	if len(packed) > int(s.width) {
		return fmt.Errorf("packed payload too large: %d", len(packed))
	}
	s.ensure(pos)
	s.length[pos] = uint8(len(packed))
	base := pos * int(s.width)
	copy(s.payload[base:base+len(packed)], packed)
	return nil
}

// Packed returns the compact payload for a trie position.
func (s *patternStore) Packed(pos int) ([]byte, bool) {
	if pos < 0 || pos >= len(s.length) {
		return nil, false
	}
	n := s.length[pos]
	if n == absentPayload {
		return nil, false
	}
	base := pos * int(s.width)
	return s.payload[base : base+int(n)], true
}

// MergeInto merges payload at trie position pos into dst at absolute offset at.
func (s *patternStore) MergeInto(pos int, at int, dst []int) []int {
	packed, ok := s.Packed(pos)
	if !ok {
		return dst
	}
	for _, b := range packed {
		rel := int(b >> 4)
		val := int(b & 0x0F)
		abs := at + rel
		for abs >= len(dst) {
			dst = append(dst, 0)
		}
		if val > dst[abs] {
			dst[abs] = val
		}
	}
	return dst
}
