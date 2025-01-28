package grid

import (
	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

// MockAtom is a mock Atom for testing.
type MockAtom struct {
	data     []byte
	hashes   []multihash.Multihash
	lastHash int
}

// NewMockAtom returns a new MockAtom.
func NewMockAtom(data []byte) *MockAtom {
	return &MockAtom{data: data}
}

// HashMRU returns the most recently computed multihash of the Atom.
func (a *MockAtom) HashMRU() multihash.Multihash {
	return a.hashes[a.lastHash]
}

// HashAdd adds and returns the multihash of the Atom given a
// multihash code.  If the Atom already has a hash using the same
// algorithm, it will replace the old hash with the new one.
func (a *MockAtom) HashAdd(code uint64) multihash.Multihash {
	// hash the data using multihash.Encode
	buf, err := multihash.Encode(a.data, code)
	Ck(err)
	return buf
}
