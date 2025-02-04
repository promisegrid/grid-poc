package interfaces_git

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"slices"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/fxamacker/cbor/v2"
	. "github.com/stevegt/goadapt"
)

/*
Imports might work like this in a PromiseGrid-based language:

import (
	dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f as bar
)

bar.foo()
*/

const (
	BlobTagName   = "blob"
	TreeTagName   = "tree"
	CommitTagName = "commit"
)

// Putting mocks and tests in one file like this is not idiomatic Go.
// We're doing it here while getting thoughts in order. All of this
// will be moved to other files in the future.

// This file is a rough approximation of git semantics for demo and
// discussion purposes.  It is not intended to be compatible with
// git at all.  For instance, it uses CBOR serialization instead of
// the git object serialization format.

// Hash returns the hash of the object as a hex string.  The hash is
// a sha-256 hash of the entire CBOR serialized object.  We use RFC
// 8949 section 4.2.1 core deterministic encoding for CBOR serialization.
func Hash(obj Object) (strhash string) {
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

	// XXX Reuse this encoding mode. It is safe for concurrent use.
	// XXX ensure we're doing all the things we need to do for deterministic encoding
	// per section 4.2.1 of RFC 8949
	opts := cbor.CoreDetEncOptions() // use preset options as a starting point
	// opts.Time = cbor.TimeUnix          // change any settings if needed
	em, err := opts.EncMode() // create an immutable encoding mode
	Ck(err)

	typ := obj.Type()
	// convert typ string to an integer
	typTag := binary.BigEndian.Uint32([]byte(typ))
	// marshal the object into a CBOR byte slice
	payload, err := em.Marshal(obj)
	Ck(err)

	// Create nested CBOR tags per RFC 9277 section 2.2
	msg := cbor.Tag{
		Number: 55799, // CBOR Wrapped tag (0xd9d9f7)
		Content: cbor.Tag{
			// Number:  1735551332, // "grid" tag (0x67726964)
			Number:  uint64(typTag),
			Content: payload,
		},
	}

	buf, err = em.Marshal(msg)
	Ck(err)
	return
}

// Decode returns the object decoded from the CBOR buffer.
func Decode(buf []byte) (obj Object, err error) {
	// create a new decoder
	opts := cbor.DecOptions{}
	decoder, err := opts.DecMode().NewDecoder(buf)

	// expect the file content to be a CBOR wrapped tag
	msg := cbor.Tag{}
	err = decoder.Decode(&msg)
	Ck(err)
	spew.Dump(msg)
	// expect the wrapped tag to contain a CBOR tag
	typ := cbor.Tag{}
	err = decoder.Decode(&typ)
	Ck(err)
	// expect the CBOR tag to contain the object
	err = decoder.Decode(obj)
	Ck(err)
	return
}

// PrintDiag prints a human-readable diagnostic string for the given CBOR
// buffer.
func PrintDiag(buf []byte) {
	// Diagnose the CBOR data
	dm, err := cbor.DiagOptions{
		ByteStringText:         true,
		ByteStringEmbeddedCBOR: true,
	}.DiagMode()
	Ck(err)
	diagnosis, err := dm.Diagnose(buf)
	Ck(err)
	Pl(diagnosis)
}

// MockBlob is a test implementation of the Blob interface.
type MockBlob struct {
	Content []byte
}

// NewMockBlob creates a new MockBlob given a type and content.
func NewMockBlob(content []byte) (obj *MockBlob) {
	obj = &MockBlob{
		Content: content,
	}
	return
}

// Type returns the type of the blob.
func (obj *MockBlob) Type() string {
	return "blob"
}

// GetSize returns the size of the blob in bytes.
func (obj *MockBlob) GetSize() int {
	return len(obj.Content)
}

// GetContent returns the content of the blob.
func (obj *MockBlob) GetContent() []byte {
	return obj.Content
}

// TestBlobHash tests the Hash method of the Blob interface.
func TestBlobHash(t *testing.T) {
	// Create a new Blob
	obj := NewMockBlob([]byte("Hello, World!"))
	buf, err := Encode(obj)
	Ck(err)
	PrintDiag(buf)
	// Test the Hash method
	want := "bfb32213ffd89fc1cdb406c8835ec68bd61a5a6c637bd5829b8ee0d7a173fa90"
	Tassert(t, want == Hash(obj), "Expected %s, got %s", want, Hash(obj))
	/*
		if Hash(obj) == want {
			t.Errorf("Expected %s, got %s", want, Hash(obj))
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
func (store *MockStore) Put(obj Object) (hash string, err error) {
	hash = Hash(obj)
	fn := store.dir + "/" + hash
	fh, err := os.Create(fn)
	Ck(err)
	defer fh.Close()
	buf, err := Encode(obj)
	Ck(err)
	_, err = fh.Write(buf)
	Ck(err)
	return
}

// Get retrieves an object from disk.
func (store *MockStore) Get(hash string) (obj Object, err error) {
	fn := store.dir + "/" + hash
	// read the entire file into buf
	// XXX
}

// TestStore tests the Store interface.
func TestStore(t *testing.T) {
	// Create a new Store
	store := NewMockStore("/tmp/mockstore")
	// Create a new Object
	obj := NewMockBlob([]byte("Hello, World!"))
	// Store the object
	hash1, err := store.Put(obj)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	// Retrieve the object
	obj2 := &MockBlob{}
	err = store.Get(hash1, obj2)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	hash2 := Hash(obj2)
	Tassert(t, hash1 == hash2, "Expected %s, got %s", hash1, hash2)
}

// TestBlob tests the Blob interface.
func TestBlob(t *testing.T) {
	// Create a new Blob
	blob := NewMockBlob([]byte("Hello, World!"))
	// Test the Type method
	Tassert(t, blob.Type() == "blob", "Expected blob, got %s", blob.Type())
	// Test the Content method
	Tassert(t, string(blob.GetContent()) == "Hello, World!", "Expected Hello, World!, got %s", string(blob.GetContent()))
	// Test the Size method
	Tassert(t, blob.GetSize() == 13, "Expected 13, got %d", blob.GetSize())
	// Test the Hash method
	want := "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d"
	Tassert(t, want == Hash(blob), "Expected %s, got %s", want, Hash(blob))
}

// MockTree is a test implementation of the Tree interface.
type MockTree struct {
	Entries []*MockEntry
}

// NewMockTree creates a new MockTree
func NewMockTree() (tree *MockTree) {
	tree = &MockTree{
		Entries: make([]*MockEntry, 0),
	}
	return
}

// Type returns the type of the object.
func (tree *MockTree) Type() string {
	return "tree"
}

// AddEntry adds an entry to the tree.
func (tree *MockTree) AddEntry(entry *MockEntry) {
	tree.Entries = append(tree.Entries, entry)
}

// GetEntries returns the entries in the tree.
func (tree *MockTree) GetEntries() []*MockEntry {
	return tree.Entries
}

// String returns a string representation of the tree.
func (tree *MockTree) String() (str string) {
	// sort the entries by name using SortFunc with a comparator
	// XXX replace sort with serialization/deserialization
	slices.SortFunc(tree.Entries, func(a, b *MockEntry) int {
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
func NewMockEntry(name, hash, mode string) (entry *MockEntry) {
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
	Tassert(t, tree.Type() == "tree", "Expected tree, got %s", tree.Type())
	// Test the AddEntry method
	entry := NewMockEntry("file.txt", "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d", "100644")
	tree.AddEntry(entry)
	Tassert(t, len(tree.GetEntries()) == 1, "Expected 1, got %d", len(tree.GetEntries()))
	// Add another entry
	entry = NewMockEntry("file2.txt", "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d", "100644")
	tree.AddEntry(entry)
	// Test the Hash
	want := "651a072cec2b04c195dbc90b208b8acd37319cfcafdb3d9ca9b6a11c32fda481"
	Tassert(t, want == Hash(tree), "Expected %s, got %s", want, Hash(tree))
	// Test the String method
	want = "tree\n100644 2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d file.txt\n100644 2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d file2.txt\n"
	Tassert(t, want == tree.String(), "Expected %s, got %s", want, tree.String())
	Pl(tree.String())

	// Test storing and retrieving a tree
	store := NewMockStore("/tmp/mockstore")
	hash1, err := store.Put(tree)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	tree2 := &MockTree{}
	err = store.Get(hash1, tree2)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	hash2 := Hash(tree2)
	if hash1 != hash2 {
		// show the difference
		diag1 := tree.String()
		diag2 := tree2.String()
		Pl(diag1)
		Pl(diag2)
		t.Errorf("Expected %s, got %s", hash1, hash2)
	}

}

// Commit is an interface for a commit object in a Git repository.
type Commit interface {
	Object
	GetTree() string
	GetParents() []string
	GetAuthor() string
	GetMessage() string
}
