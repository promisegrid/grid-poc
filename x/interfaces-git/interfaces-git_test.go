package interfaces_git

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"slices"
	"testing"

	codec "github.com/stevegt/grid-poc/x/cbor-codec"

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
func Hash(buf []byte) (strhash string) {
	binhash := sha256.Sum256(buf)
	strhash = hex.EncodeToString(binhash[:])
	return
}

// XXX use these
const (
	BlobTagName   = "blob"
	TreeTagName   = "tree"
	CommitTagName = "commit"
)

// setupCodecForGit creates a codec instance and registers the appropriate types.
func setupCodecForGit() *codec.Codec {
	config := codec.CodecConfig{
		EncOptions: cbor.CoreDetEncOptions(),
		DecOptions: cbor.DecOptions{},
	}
	c, err := codec.NewCodec(config)
	Ck(err)

	return c
}

// PrintDiag prints a human-readable diagnostic string for the given CBOR buffer.
func PrintDiag(buf []byte) {
	dm, err := cbor.DiagOptions{
		ByteStringText:         true,
		ByteStringEmbeddedCBOR: true,
	}.DiagMode()
	Ck(err)
	diagnosis, err := dm.Diagnose(buf)
	Ck(err)
	Pl(diagnosis)
}

// PrintSpew prints a human-readable spew dump of the given object.
func PrintSpew(buf []byte, c *codec.Codec) {
	tagNum, inner, err := c.DecodeTag(buf)
	Ck(err)
	Pl("tagNum:", tagNum)
	if tagNum == 0 {
		// no tag, decode as inner object
		inner = buf
	}
	var obj interface{}
	err = c.DecodeRaw(inner, &obj)
	spew.Dump(obj)
	spew.Dump(err)
}

// MockBlob is a test implementation of the Blob interface.
type MockBlob struct {
	Content []byte
}

// NewMockBlob creates a new MockBlob with the given content.
func NewMockBlob(content []byte) *MockBlob {
	return &MockBlob{
		Content: content,
	}
}

// Type returns the type of the blob.
func (b *MockBlob) Type() string {
	return "blob"
}

// GetSize returns the size of the blob in bytes.
func (b *MockBlob) GetSize() int {
	return len(b.Content)
}

// GetContent returns the content of the blob.
func (b *MockBlob) GetContent() []byte {
	return b.Content
}

/*
// TestBlobHash tests the Hash method of the Blob interface.
func TestBlobHash(t *testing.T) {
	c := setupCodecForGit()
	// Create a new Blob.
	obj := NewMockBlob([]byte("Hello, World!"))
	buf, err := c.Encode(obj)
	Ck(err)
	PrintDiag(buf)
	PrintSpew(buf, c)
	// Test the Hash method
	want := "f933ebe84766bd1227671108c66a497e31483eabf11d741bc147cfcdf67fadb2"
	Tassert(t, want == Hash(obj, c), "Expected %s, got %s", want, Hash(obj, c))
}
*/

// MockStore is a test implementation of the Store interface.
type MockStore struct {
	dir   string
	codec *codec.Codec
}

// NewMockStore creates a new MockStore with the provided directory and codec.
func NewMockStore(dir string, c *codec.Codec) Store {
	return &MockStore{
		dir:   dir,
		codec: c,
	}
}

// Put stores an object on disk and returns the hash of the object.
func (store *MockStore) Put(obj Object) (strHash string, err error) {
	c := store.codec
	// get tag number from object type name
	tagNum := codec.StringToNum(obj.Type())
	// encode the object
	buf, err := c.Encode(tagNum, obj)
	Ck(err)
	// hash the encoded object
	strHash = Hash(buf)
	// store the object on disk
	fn := store.dir + "/" + strHash
	fh, err := os.Create(fn)
	Ck(err)
	defer fh.Close()
	_, err = fh.Write(buf)
	Ck(err)
	// return the hash
	return
}

// Get retrieves an object from disk and decodes it into the returned object
func (store *MockStore) Get(hash string) (obj Object, err error) {
	fn := store.dir + "/" + hash
	buf, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	c := store.codec
	// get tag number
	tagNum, inner, err := c.DecodeTag(buf)
	Ck(err)
	// get tag name
	tagName := codec.NumToString(tagNum)
	// decode inner object
	switch tagName {
	case BlobTagName:
		var decoded MockBlob
		err = c.DecodeRaw(inner, &decoded)
		Ck(err)
		return &decoded, nil
	case TreeTagName:
		var decoded MockTree
		err = c.DecodeRaw(inner, &decoded)
		Ck(err)
		return &decoded, nil
		/*
			case CommitTagName:
				var decoded MockCommit
				err = c.DecodeRaw(inner, &decoded)
				Ck(err)
				return &decoded, nil
		*/
	default:
		return nil, fmt.Errorf("Unknown tag name: %s", tagName)
	}
}

// verify re-encodes the given object and compares the hash to the
// given hash, returning the new hash and true if they match.
func verify(obj Object, oldHash string, c *codec.Codec) (newHash string, ok bool) {
	tagNum := codec.StringToNum(obj.Type())
	buf, err := c.Encode(tagNum, obj)
	if err != nil {
		fmt.Println(err)
		ok = false
		return
	}
	newHash = Hash(buf)
	ok = newHash == oldHash
	return
}

// TestStore tests the Store interface.
func TestStore(t *testing.T) {
	c := setupCodecForGit()
	// Create a new Store.
	store := NewMockStore("/tmp/mockstore", c)
	// Create a new Object.
	obj := NewMockBlob([]byte("Hello, World!"))
	// Store the object
	hash1, err := store.Put(obj)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	// Retrieve the object
	obj2, err := store.Get(hash1)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	// Verify the object
	hash2, ok := verify(obj2, hash1, c)
	if !ok {
		t.Errorf("Expected %s, got %s", hash1, hash2)
	}
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
}

// MockTree is a test implementation of the Tree interface.
type MockTree struct {
	Entries []*MockEntry
}

// NewMockTree creates a new MockTree.
func NewMockTree() *MockTree {
	return &MockTree{
		Entries: make([]*MockEntry, 0),
	}
}

// Type returns the type of the tree.
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
func (tree *MockTree) String() string {
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
	str := "tree\n"
	for _, entry := range tree.Entries {
		str += entry.GetMode() + " " + entry.GetHash() + " " + entry.GetName() + "\n"
	}
	return str
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

// MockEntry is a test implementation of an Entry in a tree.
type MockEntry struct {
	Name string
	Hash string
	Mode string
}

// NewMockEntry creates a new MockEntry with the given name, hash, and mode.
func NewMockEntry(name, hash, mode string) *MockEntry {
	return &MockEntry{
		Name: name,
		Hash: hash,
		Mode: mode,
	}
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
	c := setupCodecForGit()
	// Create a new Tree.
	tree := NewMockTree()
	// Test the Type method.
	Tassert(t, tree.Type() == "tree", "Expected tree, got %s", tree.Type())
	// Test the AddEntry method.
	entry := NewMockEntry("file.txt", "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d", "100644")
	tree.AddEntry(entry)
	Tassert(t, len(tree.GetEntries()) == 1, "Expected 1, got %d", len(tree.GetEntries()))
	// Add another entry.
	entry = NewMockEntry("file2.txt", "2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d", "100644")
	tree.AddEntry(entry)
	Pl(tree.String())
	// verify
	want := "744f6b1da40c23135986a143ef9cdedac60387df468f708da590cfb73537f8b3"
	got, ok := verify(tree, want, c)
	Tassert(t, ok, "Expected %s, got %s", want, got)
	// Test the String method.
	wantStr := "tree\n100644 2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d file.txt\n100644 2bbef151425ac7b6e79482589fd28d21bd852422bc0ca70f26a8f8792e8f934d file2.txt\n"
	Tassert(t, wantStr == tree.String(), "Expected %s, got %s", wantStr, tree.String())

	// Test storing and retrieving a tree.
	store := NewMockStore("/tmp/mockstore", c)
	hash1, err := store.Put(tree)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	tree2intf, err := store.Get(hash1)
	tree2 := tree2intf.(*MockTree)
	Tassert(t, err == nil, "Expected nil, got %v", err)
	hash2, ok := verify(tree2, hash1, c)
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
