package hyphenate

import (
	"io"
	"testing"
)

type slicePatternReader struct {
	entries []Pattern
	index   int
}

func (r *slicePatternReader) Next() ([]rune, []int, error) {
	if r.index >= len(r.entries) {
		return nil, nil, io.EOF
	}
	entry := r.entries[r.index]
	r.index++
	return entry.Sequence, entry.Weights, nil
}

type sliceExceptionReader struct {
	entries []struct {
		word      string
		positions []int
	}
	index int
}

func (r *sliceExceptionReader) Next() (string, []int, error) {
	if r.index >= len(r.entries) {
		return "", nil, io.EOF
	}
	entry := r.entries[r.index]
	r.index++
	return entry.word, entry.positions, nil
}

func TestPatternReaderAPI(t *testing.T) {
	dict, err := LoadPatternReader("stream-patterns", &slicePatternReader{
		entries: []Pattern{{
			Sequence: []rune("für"),
			Weights:  []int{0, 0, 1},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if h := dict.HyphenationString("fürung"); h != "fü-rung" {
		t.Fatalf("fürung should be fü-rung, is %s", h)
	}
}

func TestPatternListAPI(t *testing.T) {
	dict, err := LoadPatternReader("list-patterns", &slicePatternReader{
		entries: []Pattern{{
			Sequence: []rune("für"),
			Weights:  []int{0, 0, 1},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if h := dict.HyphenationString("fürung"); h != "fü-rung" {
		t.Fatalf("fürung should be fü-rung, is %s", h)
	}
}

func TestExceptionReaderAPI(t *testing.T) {
	dict, err := LoadPatternReader("stream-exceptions", &slicePatternReader{})
	if err != nil {
		t.Fatal(err)
	}
	dict.LoadExceptionReader(&sliceExceptionReader{
		entries: []struct {
			word      string
			positions []int
		}{
			{
				word:      "table",
				positions: []int{0, 0, 1, 0, 0},
			},
		},
	})
	if h := dict.HyphenationString("table"); h != "ta-ble" {
		t.Fatalf("table should be ta-ble, is %s", h)
	}
}

func TestPatternTrieStats(t *testing.T) {
	dict, err := LoadPatternReader("stats", &slicePatternReader{
		entries: []Pattern{
			{Sequence: []rune("ab"), Weights: []int{0, 1}},
			{Sequence: []rune("abc"), Weights: []int{0, 1, 0}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	backend, used, total, maxStateID, fill := dict.PatternTrieStats()
	if backend != "dat" {
		t.Fatalf("expected dat backend, got %s", backend)
	}
	if used <= 0 || total <= 0 {
		t.Fatalf("expected positive slot counts, got used=%d total=%d", used, total)
	}
	if maxStateID <= 0 {
		t.Fatalf("expected positive maxStateID, got %d", maxStateID)
	}
	if fill <= 0 || fill > 1 {
		t.Fatalf("expected fill ratio in (0,1], got %f", fill)
	}
}
