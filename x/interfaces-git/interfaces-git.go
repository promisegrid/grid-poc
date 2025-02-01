package interfaces_git

// . "github.com/stevegt/goadapt"

// XXX Putting all of the interfaces in one file like this is not
// idiomatic Go.  We're doing it here while getting thoughts in order.
// All of these interfaces will be moved to their caller's files in
// the future.

// Object is an interface for objects that can be stored in a Git
// repository.
type Object interface {
	// GetHash returns the hash of the object as a hex string.  The
	// hash is the sha256 hash of the object's encoded CBOR.
	GetHash() string
	// GetType returns the type of the object, e.g. "blob", "tree", "commit".
	GetType() string
	// GetSize returns the size of the object in bytes.
	GetSize() int
	// Encode returns the object encoded in deterministic CBOR format
	// per RFC 8949 section 4.2.1.
	Encode() ([]byte, error)
}

// Store is an interface for storing objects on disk.
type Store interface {
	// Put stores an object on disk.
	Put(Object) error
	// Get retrieves an object from disk.
	Get(string) (Object, error)
}

// Blob is an interface for a blob object in a Git repository.
type Blob interface {
	Object
	// GetContent returns the content of the blob.
	GetContent() []byte
}

// Tree is an interface for a tree object in a Git repository.
type Tree interface {
	Object
	AddEntry(Entry)
	GetEntries() []Entry
	String() string
}

// Entry is an interface for an entry in a tree object in a Git
// repository.
type Entry interface {
	GetName() string
	GetHash() string
	GetMode() string
}

// Commit is an interface for a commit object in a Git repository.
type Commit interface {
	Object
	GetTree() string
	GetParents() []string
	GetAuthor() string
	GetMessage() string
}
