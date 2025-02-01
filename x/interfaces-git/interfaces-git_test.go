package interfaces_git

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"slices"
	"testing"

	"github.com/fxamacker/cbor/v2"
	. "github.com/stevegt/goadapt"
)

/*
e.g. in PromiseGrid:
import (
	dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f as bar
)

bar.foo()
*/

// This file is a rough approximation of git semantics for demo and
// discussion purposes.  It is not intended to be compatible with
// git at all.  For instance, it uses CBOR serialization instead of
// the git object serialization format.

// MockBlob is a test implementation of the Object interface.
type MockBlob struct {
	Content []byte
	Type    string
}

// NewMockBlob creates a new MockBlob given a type and content.
func NewMockBlob(content []byte) (obj *MockBlob) {
	obj = &MockBlob{
		Content: content,
		Type:    "blob",
	}
	return
}

// GetType returns the type of the object.
func (obj *MockBlob) GetType() string {
	return obj.Type
}

// GetContent returns the content of the object.
func (obj *MockBlob) GetContent() []byte {
	return obj.Content
}

// GetSize returns the size of the object in bytes.
func (obj *MockBlob) GetSize() int {
	return len(obj.Content)
}

// GetHash returns the hash of the object as a hex string.  The hash is
// a sha-256 hash of the entire CBOR serialized object.  We use RFC
// 8949 core deterministic encoding for CBOR serialization.
func (obj *MockBlob) GetHash() (strhash string) {
	buf, err := Encode(obj)
	Ck(err)
	binhash := sha256.Sum256(buf)
	strhash = hex.EncodeToString(binhash[:])
	return
}

// Encode returns the object encoded in deterministic CBOR format
// per RFC 8949 section 4.2.1.
func Encode(obj Object) (buf []byte, err error) {
	defer Return(&err)
	// XXX ensure we're doing all the things we need to do for deterministic encoding
	// per section 4.2.1 of RFC 8949
	opts := cbor.CoreDetEncOptions() // use preset options as a starting point
	// opts.Time = cbor.TimeUnix          // change any settings if needed
	// XXX Reuse the encoding mode. It is safe for concurrent use.
	em, err := opts.EncMode() // create an immutable encoding mode
	Ck(err)
	buf, err = em.Marshal(obj)
	Ck(err)
	return
}

// TestObjectHash tests the Hash method of the Object interface.
func TestObjectHash(t *testing.T) {
	// Create a new Object
	obj := NewMockBlob([]byte("Hello, World!"))
	// Test the Hash method
	want := "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d"
	Tassert(t, want == obj.GetHash(), "Expected %s, got %s", want, obj.GetHash())
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

// Put stores an object on disk.
func (store *MockStore) Put(obj Object) (err error) {
	fn := store.dir + "/" + obj.GetHash()
	fh, err := os.Create(fn)
	Ck(err)
	defer fh.Close()
	/*
		// write type as the first line
		_, err = fh.Write([]byte(obj.GetType() + "\n"))
		// write content
		_, err = fh.Write(obj.GetContent())
		Ck(err)
	*/
	buf, err := Encode(obj)
	Ck(err)
	_, err = fh.Write(buf)
	Ck(err)
	return
}

// Get retrieves an object from disk.
func (store *MockStore) Get(hash string) (obj Object, err error) {
	fn := store.dir + "/" + hash
	fh, err := os.Open(fn)
	Ck(err)
	defer fh.Close()

	/*
		// type is the first line -- read bytes until newline
		// XXX this all gets replaced by CBOR serialization
		var typ []byte
		b := make([]byte, 1)
		for i := 0; i < 99; i++ {
			_, err = fh.Read(b)
			Ck(err)
			if b[0] == '\n' {
				break
			}
			typ = append(typ, b[0])
		}

		// read the rest of the file
		// get the size of the file
		fi, err := fh.Stat()
		Ck(err)
		// subtract the size of the type line
		size := fi.Size() - int64(len(typ)) - 1
		// read the content into a buffer
		content := make([]byte, size)
		_, err = fh.Read(content)
		Ck(err)
		obj = NewMockBlob(string(typ), content)
	*/

	decoder := cbor.NewDecoder(fh)
	obj = &MockBlob{}
	err = decoder.Decode(obj)
	Ck(err)

	return
}

// TestStore tests the Store interface.
func TestStore(t *testing.T) {
	// Create a new Store
	store := NewMockStore("/tmp/mockstore")
	// Create a new Object
	obj := NewMockBlob([]byte("Hello, World!"))
	// Store the object
	err := store.Put(obj)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	// Retrieve the object
	raw2, err := store.Get(obj.GetHash())
	obj2 := raw2.(*MockBlob)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	Tassert(t, obj.GetHash() == obj2.GetHash(), "Expected %s, got %s", obj.GetHash(), obj2.GetHash())
	Tassert(t, string(obj.GetContent()) == string(obj2.GetContent()), "Expected %s, got %s", string(obj.GetContent()), string(obj2.GetContent()))
}

// TestBlob tests the Blob interface.
func TestBlob(t *testing.T) {
	// Create a new Blob
	blob := NewMockBlob([]byte("Hello, World!"))
	// Test the Type method
	Tassert(t, blob.GetType() == "blob", "Expected blob, got %s", blob.GetType())
	// Test the Content method
	Tassert(t, string(blob.GetContent()) == "Hello, World!", "Expected Hello, World!, got %s", string(blob.GetContent()))
	// Test the Size method
	Tassert(t, blob.GetSize() == 13, "Expected 13, got %d", blob.GetSize())
	// Test the Hash method
	want := "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d"
	Tassert(t, want == blob.GetHash(), "Expected %s, got %s", want, blob.GetHash())
}

// MockTree is a test implementation of the Tree interface.
type MockTree struct {
	MockBlob
	Entries []Entry
}

// NewMockTree creates a new MockTree given a list of entries.
func NewMockTree() (tree Tree) {
	tree = &MockTree{
		MockBlob: MockBlob{
			Type: "tree",
		},
	}
	return
}

// AddEntry adds an entry to the tree.
func (tree *MockTree) AddEntry(entry Entry) {
	tree.Entries = append(tree.Entries, entry)
}

// GetEntries returns the entries in the tree.
func (tree *MockTree) GetEntries() []Entry {
	return tree.Entries
}

// String returns a string representation of the tree.
func (tree *MockTree) String() (str string) {
	// sort the entries by name using SortFunc with a comparator
	slices.SortFunc(tree.Entries, func(a, b Entry) int {
		if a.GetName() == b.GetName() {
			return 0
		}
		if a.GetName() < b.GetName() {
			return -1
		}
		return 1
	})
	str = "tree\n"
	for _, entry := range tree.Entries {
		str += entry.GetMode() + " " + entry.GetHash() + " " + entry.GetName() + "\n"
	}
	return
}

// GetHash returns the hash of the tree.
func (tree *MockTree) GetHash() (strhash string) {
	// XXX this all gets replaced by CBOR serialization
	// hash the sorted entries
	binhash := sha256.Sum256([]byte(tree.String()))
	strhash = hex.EncodeToString(binhash[:])
	// return the hash
	return
}

// MockEntry is a test implementation of the Entry interface.
type MockEntry struct {
	Name string
	Hash string
	Mode string
}

// NewMockEntry creates a new MockEntry given a name, hash, and mode.
func NewMockEntry(name, hash, mode string) (entry Entry) {
	entry = &MockEntry{
		Name: name,
		Hash: hash,
		Mode: mode,
	}
	return
}

// GetName returns the name of the entry.
func (entry *MockEntry) GetName() string {
	return entry.Name
}

// GetHash returns the hash of the entry.
func (entry *MockEntry) GetHash() string {
	return entry.Hash
}

// GetMode returns the mode of the entry.
func (entry *MockEntry) GetMode() string {
	return entry.Mode
}

// TestTree tests the Tree interface.
func TestTree(t *testing.T) {
	// Create a new Tree
	tree := NewMockTree()
	// Test the Type method
	Tassert(t, tree.GetType() == "tree", "Expected tree, got %s", tree.GetType())
	// Test the AddEntry method
	entry := NewMockEntry("file.txt", "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d", "100644")
	tree.AddEntry(entry)
	Tassert(t, len(tree.GetEntries()) == 1, "Expected 1, got %d", len(tree.GetEntries()))
	// Add another entry
	entry = NewMockEntry("file2.txt", "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d", "100644")
	tree.AddEntry(entry)
	// Test the Hash method
	want := "8dacb749aa28e977c9eac0c4ac57e3d3a33ec8b684f66025ae3ea9952fd35a31"
	Tassert(t, want == tree.GetHash(), "Expected %s, got %s", want, tree.GetHash())
	// Test the String method
	want = "tree\n100644 2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d file.txt\n100644 2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d file2.txt\n"
	Tassert(t, want == tree.String(), "Expected %s, got %s", want, tree.String())
	Pl(tree.String())

	// Test storing and retrieving a tree
	store := NewMockStore("/tmp/mockstore")
	err := store.Put(tree)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	tree2, err := store.Get(tree.GetHash())
	Tassert(t, err == nil, "Expected nil, got %v", err)
	Tassert(t, tree.GetHash() == tree2.GetHash(), "Expected %s, got %s", tree.GetHash(), tree2.GetHash())

}
