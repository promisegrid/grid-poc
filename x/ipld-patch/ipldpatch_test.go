package ipldpatch

import (
	"testing"
)

func TestRunDemo(t *testing.T) {
	// Capture potential panics
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Demo panicked: %v", r)
		}
	}()

	RunDemo()
}
