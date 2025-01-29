package grid

import (
	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

// MockAtom is a mock Atom for testing.
type MockAtom struct {
	data     []byte
	hashes   map[uint64]multihash.Multihash
	lastHash uint64
}

// NewMockAtom returns a new MockAtom.
func NewMockAtom(data []byte) *MockAtom {
	return &MockAtom{data: data}
}

// HashMRU returns the most recently computed multihash of the Atom.
func (a *MockAtom) HashMRU() multihash.Multihash {
	hash, ok := a.hashes[a.lastHash]
	if !ok {
		return nil
	}
	return hash
}

// HashAdd adds and returns the multihash of the Atom given a
// multihash code.  If the Atom already has a hash using the same
// algorithm, it will replace the old hash with the new one.
func (a *MockAtom) HashAdd(code uint64) multihash.Multihash {
	// hash the data using multihash.Encode
	buf, err := multihash.Encode(a.data, code)
	Ck(err)
	a.hashes[code] = buf
	return buf
}

/*
// HashAddName adds and returns the multihash of the Atom given a
// multihash name.  If the Atom already has a hash using the same
// algorithm, it will replace the old hash with the new one.
func (a *MockAtom) HashAddName(name string) multihash.Multihash {
	// hash the data using multihash.Encode
	buf, err := multihash.EncodeName(a.data, name)
	Ck(err)
	code, ok := multihash.Names[code]
	if !ok {
		return nil
	}
	// a.hashes[buf.Code] = buf
	return buf
}

// HashGet returns the multihash of the Atom given a multihash code.
func (a *MockAtom) HashGet(code uint64) multihash.Multihash {
	hash, ok := a.hashes[code]
	if !ok {
		a.HashAdd(code)
		hash = a.hashes[code]
	}
	return hash
}

// HashGetName returns the multihash of the Atom given a multihash name.
func (a *MockAtom) HashGetName(name string) multihash.Multihash {
	code, err := multihash.LookupCode(name)
	Ck(err)
	hash, ok := a.hashes[code]
	if !ok {
		return nil
	}
	return hash
}
*/
