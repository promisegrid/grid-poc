package interfaces_git

// . "github.com/stevegt/goadapt"

// XXX Putting all of the interfaces in one file like this is not
// idiomatic Go.  We're doing it here while getting thoughts in order.
// All of these interfaces will be moved to their caller's files in
// the future.

// Object is an interface for objects that can be stored in a Git
// repository.
type Object interface {
	// GetHash returns the hash of the object as a hex string.
	GetHash() string
	// GetType returns the type of the object, e.g. "blob", "tree", "commit".
	GetType() string
	// Content returns the content of the object.  This would be the
	// content of a blob, the list of entries in a tree, or the commit
	// message in a commit.
	Content() []byte
	// Size returns the size of the object in bytes.
	Size() int
}

// Store is an interface for storing objects on disk.
type Store interface {
	// Store stores an object on disk.
	Store(Object) error
	// Retrieve retrieves an object from disk.
	Retrieve(string) (Object, error)
}

// Blob is an interface for a blob object in a Git repository.
type Blob interface {
	Object
}

// Tree is an interface for a tree object in a Git repository.
type Tree interface {
	Object
	AddEntry(Entry)
	Entries() []Entry
	String() string
}

// Entry is an interface for an entry in a tree object in a Git
// repository.
type Entry interface {
	Name() string
	Hash() string
	Mode() string
}

// Commit is an interface for a commit object in a Git repository.
type Commit interface {
	Object
	Tree() string
	Parents() []string
	Author() string
	Message() string
}
