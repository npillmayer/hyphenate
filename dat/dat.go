package dat

// DAT is a frozen double-array trie for Liang-style hyphenation patterns.
// - Nodes/states are indices into Base/Check (0 is unused; Root is typically 1).
// - Transition: t := Base[s] + c; valid if Check[t] == s; next state is t.
// - c is a dense alphabet ID in [1..Sigma]. c==0 means "not in alphabet".
//
// Payloads:
//   - If PayloadOff[s] != 0, node s is terminal and has an associated digit sequence.
//   - Digit sequences are stored in Payload as packed nibbles (two digits per byte).
//   - Each payload record begins with a VarUint length L (number of digits),
//     followed by ceil(L/2) bytes of packed digits.
//
// Mapping:
//   - Map is a BMP mapping from UTF-16 code unit (0..65535) to dense alphabet ID.
//     0 means "not part of the pattern alphabet".
type DAT struct {
	// Root state index (commonly 1).
	Root uint32

	// Sigma is the size of the dense alphabet (maximum dense ID).
	Sigma uint16

	// Base and Check are the classic double-array.
	// Use signed ints to allow negative base if you ever choose that convention;
	// but here we keep them non-negative and use int32 for compactness.
	Base  []int32 // len == N
	Check []int32 // len == N

	// PayloadOff holds offsets into Payload for terminal nodes.
	// 0 means "no payload". Offsets are indices into Payload (byte slice).
	PayloadOff []uint32 // len == N

	// Payload is a blob of packed digit sequences.
	// Record format at offset off:
	//   - uvarint L (number of digits)
	//   - packed digits: ceil(L/2) bytes, high nibble first:
	//       byte = (d0<<4) | d1, digits are 0..9
	Payload []byte

	// MapBMP maps BMP code units to dense IDs [0..Sigma].
	// For BMP-only workflows this is the fastest mapping.
	// Memory: 65536 * 2 bytes = 128 KB per loaded language.
	MapPaged PagedMapBMP

	// Optional: MinLeft/MinRight are common hyphenation constraints.
	// Keep them here if you want the hyphenator to be configured by the DAT.
	MinLeft  uint8
	MinRight uint8
}

// NStates returns number of allocated slots/states in the arrays.
func (d *DAT) NStates() int { return len(d.Base) }

// Transition returns (nextState, ok). dense must be in [1..Sigma].
func (d *DAT) Transition(state uint32, dense uint16) (uint32, bool) {
	if int(state) >= len(d.Base) || int(state) >= len(d.Check) {
		return 0, false
	}
	t := int32(d.Base[state]) + int32(dense)
	if t <= 0 || int(t) >= len(d.Check) {
		return 0, false
	}
	if d.Check[t] != int32(state) {
		return 0, false
	}
	return uint32(t), true
}

// Dense maps a BMP code unit to a dense alphabet ID.
// Returns 0 if the code unit is not in the alphabet.
func (d *DAT) Dense(bmp uint16) uint16 { return d.MapPaged.Dense(bmp) }
