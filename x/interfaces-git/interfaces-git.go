package interfaces_git

// . "github.com/stevegt/goadapt"

// Object is an interface for objects that can be stored in a repository.
type Object interface {
	// GetType returns the type of the object, e.g. "blob", "tree", "commit".
	Type() string
}

// Store is an interface for storing objects on disk.
type Store interface {
	// Put stores an object on disk and returns the hash of the object.
	Put(Object) (string, error)
	// Get retrieves an object from disk given its hash.
	Get(string) (Object, error)
}
