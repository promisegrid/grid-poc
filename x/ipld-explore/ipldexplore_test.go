package ipldexplore

import (
	"testing"
)

func TestRunDemo(t *testing.T) {
	// Verify demo runs without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Demo panicked: %v", r)
		}
	}()

	RunDemo()
}
