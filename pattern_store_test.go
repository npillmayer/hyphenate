package hyphenate

import (
	"reflect"
	"testing"
)

func TestPatternStorePacked(t *testing.T) {
	s := newPatternStore(16)
	if err := s.Put(42, []int{0, 5, 0, 3}); err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	packed, ok := s.Packed(42)
	if !ok {
		t.Fatalf("expected payload at position 42")
	}
	want := []byte{0x15, 0x33}
	if !reflect.DeepEqual(packed, want) {
		t.Fatalf("packed mismatch: got %v, want %v", packed, want)
	}
}

func TestPatternStoreOverwrite(t *testing.T) {
	s := newPatternStore(16)
	if err := s.Put(7, []int{0, 3}); err != nil {
		t.Fatalf("first Put failed: %v", err)
	}
	if err := s.Put(7, []int{0, 9}); err != nil {
		t.Fatalf("second Put failed: %v", err)
	}
	packed, ok := s.Packed(7)
	if !ok {
		t.Fatalf("expected payload at position 7")
	}
	want := []byte{0x19}
	if !reflect.DeepEqual(packed, want) {
		t.Fatalf("packed mismatch after overwrite: got %v, want %v", packed, want)
	}
}

func TestPatternStoreMergeInto(t *testing.T) {
	s := newPatternStore(16)
	if err := s.Put(7, []int{0, 7, 3}); err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	dst := []int{0, 2, 0, 0}
	got := s.MergeInto(7, 1, dst)
	want := []int{0, 2, 7, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merge mismatch: got %v, want %v", got, want)
	}
}

func TestPatternStoreRejectsOutOfNibbleRange(t *testing.T) {
	s := newPatternStore(16)
	positions := make([]int, 17)
	positions[16] = 1
	if err := s.Put(1, positions); err == nil {
		t.Fatalf("expected out-of-range index error")
	}
}
