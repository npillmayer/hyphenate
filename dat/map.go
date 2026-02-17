package dat

// PagedMapBMP maps BMP code units (0..65535) to dense alphabet IDs (uint16).
// It's a two-level page table:
//   - Top[hi] = page index (1..NumPages), or 0 meaning "page absent".
//   - Pages is a flat array of NumPages*256 entries.
//
// Lookup is O(1) with two array reads and a couple of ops.
//
// Memory:
//   - Top: 256 * 2 = 512 bytes
//   - Each populated page: 256 * 2 = 512 bytes
//
// So if you touch, say, 40 high-byte blocks => ~20 KB for pages.
type PagedMapBMP struct {
	Top   [256]uint16 // page index (1-based); 0 means none
	Pages []uint16    // flat: NumPages*256
}

// Dense returns the dense alphabet ID for a BMP code unit.
// Returns 0 if absent.
func (m *PagedMapBMP) Dense(bmp uint16) uint16 {
	hi := bmp >> 8
	pi := m.Top[hi]
	if pi == 0 {
		return 0
	}
	base := int(pi-1) << 8 // *256
	return m.Pages[base+int(bmp&0xFF)]
}

// NumPages returns the number of allocated pages.
func (m *PagedMapBMP) NumPages() int { return len(m.Pages) >> 8 }

// EnsurePage ensures that the page for high byte hi exists.
// Returns the 1-based page index.
func (m *PagedMapBMP) EnsurePage(hi uint16) uint16 {
	pi := m.Top[hi]
	if pi != 0 {
		return pi
	}
	// allocate a new page (256 uint16 initialized to 0)
	m.Pages = append(m.Pages, make([]uint16, 256)...)
	pi = uint16(len(m.Pages) >> 8) // number of pages, 1-based index
	m.Top[hi] = pi
	return pi
}

// Set sets mapping bmp -> dense (dense may be 0 to clear).
func (m *PagedMapBMP) Set(bmp uint16, dense uint16) {
	hi := bmp >> 8
	pi := m.Top[hi]
	if pi == 0 {
		if dense == 0 {
			return
		}
		pi = m.EnsurePage(hi)
	}
	base := int(pi-1) << 8
	m.Pages[base+int(bmp&0xFF)] = dense
}
