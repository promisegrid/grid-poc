package interfaces_git

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	. "github.com/stevegt/goadapt"
)

var typ string = "blob"

// MockObject is a test implementation of the Object interface.
type MockObject struct {
	content []byte
	typ     string
}

// NewMockObject creates a new MockObject given a type and content.
func NewMockObject(typ string, content []byte) (obj *MockObject) {
	obj = &MockObject{
		content: content,
		typ:     typ,
	}
	return
}

// Type returns the type of the object.
func (obj *MockObject) Type() string {
	return obj.typ
}

// Content returns the content of the object.
func (obj *MockObject) Content() []byte {
	return obj.content
}

// Size returns the size of the object in bytes.
func (obj *MockObject) Size() int {
	return len(obj.content)
}

// Hash returns the hash of the object as a hex string.  The hash is
// a sha-256 hash of the content.
func (obj *MockObject) Hash() (strhash string) {
	binhash := sha256.Sum256(obj.content)
	strhash = hex.EncodeToString(binhash[:])
	return
}

// TestObjectHash tests the Hash method of the Object interface.
func TestObjectHash(t *testing.T) {
	// Create a new Object
	obj := NewMockObject("blob", []byte("Hello, World!"))
	// Test the Hash method
	want := "XXX"
	Tassert(t, want == obj.Hash(), "Expected %s, got %s", want, obj.Hash())
	/*
		if obj.Hash() == want {
			t.Errorf("Expected %s, got %s", want, obj.Hash())
		}
	*/
}
