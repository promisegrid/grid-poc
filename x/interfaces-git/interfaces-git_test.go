package interfaces_git

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	. "github.com/stevegt/goadapt"
)

/*
e.g. in PromiseGrid:
import (
	dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f as bar
)

bar.foo()
*/

var typ string = "blob"

// MockObject is a test implementation of the Object interface.
type MockObject struct {
	content []byte
	typ     string
}

// NewMockObject creates a new MockObject given a type and content.
func NewMockObject(typ string, content []byte) (obj Object) {
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
	want := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	Tassert(t, want == obj.Hash(), "Expected %s, got %s", want, obj.Hash())
	/*
		if obj.Hash() == want {
			t.Errorf("Expected %s, got %s", want, obj.Hash())
		}
	*/
}

// MockStore is a test implementation of the Store interface.
type MockStore struct {
	dir string
}

// NewMockStore creates a new MockStore.
func NewMockStore(dir string) (store Store) {
	store = &MockStore{
		dir: dir,
	}
	return
}

// Store stores an object on disk.
func (store *MockStore) Store(obj Object) (err error) {
	fn := store.dir + "/" + obj.Hash()
	fh, err := os.Create(fn)
	Ck(err)
	defer fh.Close()
	_, err = fh.Write(obj.Content())
	Ck(err)
	return
}

// TestStore tests the Store interface.
func TestStore(t *testing.T) {
	// Create a new Store
	store := NewMockStore("/tmp/mockstore")
	// Create a new Object
	obj := NewMockObject("blob", []byte("Hello, World!"))
	// Store the object
	err := store.Store(obj)
	// Test the Store method
	Tassert(t, err == nil, "Expected nil, got %v", err)
}
