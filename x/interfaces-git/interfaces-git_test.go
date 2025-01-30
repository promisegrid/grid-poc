package interfaces_git

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"slices"
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

// This file is a rough approximation of git semantics for demo and
// discussion purposes.  It is not intended to be compatible with
// git at all.  For instance, it uses CBOR serialization instead of
// the git object serialization format.

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

// GetType returns the type of the object.
func (obj *MockObject) GetType() string {
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

// GetHash returns the hash of the object as a hex string.  The hash is
// a sha-256 hash of the content.
func (obj *MockObject) GetHash() (strhash string) {
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

// Store stores an object on disk.
func (store *MockStore) Store(obj Object) (err error) {
	fn := store.dir + "/" + obj.GetHash()
	fh, err := os.Create(fn)
	Ck(err)
	defer fh.Close()
	// write type as the first line
	_, err = fh.Write([]byte(obj.GetType() + "\n"))
	// write content
	_, err = fh.Write(obj.Content())
	Ck(err)
	return
}

// Retrieve retrieves an object from disk.
func (store *MockStore) Retrieve(hash string) (obj Object, err error) {
	fn := store.dir + "/" + hash
	fh, err := os.Open(fn)
	Ck(err)
	defer fh.Close()

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

	obj = NewMockObject(string(typ), content)
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
	// Retrieve the object
	obj2, err := store.Retrieve(obj.GetHash())
	// Test the Retrieve method
	Tassert(t, err == nil, "Expected nil, got %v", err)
	Tassert(t, obj.GetHash() == obj2.GetHash(), "Expected %s, got %s", obj.GetHash(), obj2.GetHash())
	Tassert(t, string(obj.Content()) == string(obj2.Content()), "Expected %s, got %s", string(obj.Content()), string(obj2.Content()))
}

// MockBlob is a test implementation of the Blob interface.
type MockBlob struct {
	MockObject
	name string
}

// NewMockBlob creates a new MockBlob given a name and content.
func NewMockBlob(name string, content []byte) (blob Blob) {
	blob = &MockBlob{
		MockObject: MockObject{
			content: content,
			typ:     "blob",
		},
		name: name,
	}
	return
}

// NewBlob creates a new Blob given content.
func NewBlob(content []byte) (blob Blob) {
	blob = &MockBlob{
		MockObject: MockObject{
			content: content,
			typ:     "blob",
		},
	}
	return
}

// TestBlob tests the Blob interface.
func TestBlob(t *testing.T) {
	// Create a new Blob
	blob := NewBlob([]byte("Hello, World!"))
	// Test the Type method
	Tassert(t, blob.GetType() == "blob", "Expected blob, got %s", blob.GetType())
	// Test the Content method
	Tassert(t, string(blob.Content()) == "Hello, World!", "Expected Hello, World!, got %s", string(blob.Content()))
	// Test the Size method
	Tassert(t, blob.Size() == 13, "Expected 13, got %d", blob.Size())
	// Test the Hash method
	want := "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"
	Tassert(t, want == blob.GetHash(), "Expected %s, got %s", want, blob.GetHash())
}

// MockTree is a test implementation of the Tree interface.
type MockTree struct {
	MockObject
	entrees []Entry
}

// NewMockTree creates a new MockTree given a list of entries.
func NewMockTree() (tree Tree) {
	tree = &MockTree{
		MockObject: MockObject{
			typ: "tree",
		},
	}
	return
}

// AddEntry adds an entry to the tree.
func (tree *MockTree) AddEntry(entry Entry) {
	tree.entrees = append(tree.entrees, entry)
}

// Entries returns the entries in the tree.
func (tree *MockTree) Entries() []Entry {
	return tree.entrees
}

// String returns a string representation of the tree.
func (tree *MockTree) String() (str string) {
	// sort the entries by name using SortFunc with a comparator
	slices.SortFunc(tree.entrees, func(a, b Entry) int {
		if a.Name() == b.Name() {
			return 0
		}
		if a.Name() < b.Name() {
			return -1
		}
		return 1
	})
	str = "tree\n"
	for _, entry := range tree.entrees {
		str += entry.Mode() + " " + entry.Hash() + " " + entry.Name() + "\n"
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
	name string
	hash string
	mode string
}

// NewMockEntry creates a new MockEntry given a name, hash, and mode.
func NewMockEntry(name, hash, mode string) (entry Entry) {
	entry = &MockEntry{
		name: name,
		hash: hash,
		mode: mode,
	}
	return
}

// Name returns the name of the entry.
func (entry *MockEntry) Name() string {
	return entry.name
}

// Hash returns the hash of the entry.
func (entry *MockEntry) Hash() string {
	return entry.hash
}

// Mode returns the mode of the entry.
func (entry *MockEntry) Mode() string {
	return entry.mode
}

// TestTree tests the Tree interface.
func TestTree(t *testing.T) {
	// Create a new Tree
	tree := NewMockTree()
	// Test the Type method
	Tassert(t, tree.GetType() == "tree", "Expected tree, got %s", tree.GetType())
	// Test the AddEntry method
	entry := NewMockEntry("file.txt", "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f", "100644")
	tree.AddEntry(entry)
	Tassert(t, len(tree.Entries()) == 1, "Expected 1, got %d", len(tree.Entries()))
	// Add another entry
	entry = NewMockEntry("file2.txt", "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f", "100644")
	tree.AddEntry(entry)
	// Test the Hash method
	want := "2f3cfcd579e6319b7ee80bca4581654c903fc49d23824898070ead7758c9f691"
	Tassert(t, want == tree.GetHash(), "Expected %s, got %s", want, tree.GetHash())
	// Test the String method
	want = "tree\n100644 dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f file.txt\n100644 dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f file2.txt\n"
	Tassert(t, want == tree.String(), "Expected %s, got %s", want, tree.String())
	Pl(tree.String())
}
