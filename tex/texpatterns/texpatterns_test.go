package texpatterns

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestPatternReader(t *testing.T) {
	src := strings.NewReader(`\message{test-id}
\patterns{
fü1r
}`)
	r := NewPatternReader(src)
	seq, weights, err := r.Next()
	if err != nil {
		t.Fatalf("Next failed: %v", err)
	}
	if string(seq) != "für" {
		t.Fatalf("sequence mismatch: got %q", string(seq))
	}
	if !reflect.DeepEqual(weights, []int{0, 0, 1}) {
		t.Fatalf("weights mismatch: got %v", weights)
	}
	if r.Identifier() != "test-id" {
		t.Fatalf("identifier mismatch: %q", r.Identifier())
	}
	_, _, err = r.Next()
	if err != io.EOF {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}
