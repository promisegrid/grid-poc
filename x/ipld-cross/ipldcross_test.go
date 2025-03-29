package ipldpath

import (
	"testing"
)

func TestRunDemo(t *testing.T) {
	// Verify the demo runs without errors
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Cross-block traversal demo failed: %v", r)
		}
	}()
	
	RunDemo()
}
