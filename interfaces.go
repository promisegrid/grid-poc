package grid

import (
	"time"

	"github.com/multiformats/go-multihash"
	// . "github.com/stevegt/goadapt"
)

// XXX Putting all of the interfaces in one file like this is not
// idiomatic Go.  We're doing it here while getting thoughts in order.
// All of these interfaces will be moved to their caller's files in
// the future.

// Atom is the fundamental unit of data in a grid.  It is a single
// blob of data along with its hash(es).  An Atom is roughly analogous
// to an object in .git/objects.
type Atom interface {
	// Hash returns the most recently computed multihash of the Atom.
	HashMRU() multihash.Multihash
	// HashAdd adds and returns the multihash of the Atom given a
	// multihash code.  If the Atom already has a hash using the same
	// algorithm, it will replace the old hash with the new one.
	HashAdd(uint64) multihash.Multihash
	// HashAddName adds and returns the multihash of the Atom given a
	// multihash name.  If the Atom already has a hash using the same
	// algorithm, it will replace the old hash with the new one.
	HashAddName(string) multihash.Multihash
	// HashGet returns the multihash of the Atom given a multihash
	// code.  If the Atom does not have a hash using the given
	// algorithm, it will return nil.
	HashGet(uint64) multihash.Multihash
	// HashGetName returns the multihash of the Atom given a
	// multihash name.  If the Atom does not have a hash using the
	// given algorithm, it will return nil.
	HashGetName(string) multihash.Multihash
	// Data returns the data of the Atom.
	Data() []byte
}

// Function is a type of Atom that is used for all grid transitions on
// a world line.  It takes zero or more states as input and returns
// zero or more states as output.  The input states are the prior
// states of the output states.  The output states are the next states
// of the input states.
//
// A Function is roughly analogous to a patch in git.  It describes
// how the states changed.
type Function interface {
	Atom
	// Apply applies the function to the input states and returns the
	// output states.
	Apply(...State) []State
}

// State is a type of Atom that describes some part of the universe at
// an instant in time.  A State has a timestamp, one or more parent
// states, and a transition function that describes how the state
// changed from the prior state(s).
//
// A State is roughly analogous to a commit in git.  The parent states
// are like the parent commits of a commit.  The transition function
// is like a patch that describes how the files in the commit changed
// from the parent commit.  The timestamp is like the commit date.
type State interface {
	Atom
	Time() time.Time
	Function() Function
	Parents() []State
	Siblings() []State
}

// Promise is a type of Atom that is used for all grid messages.  It
// has a State and a Value.  The State is a list of zero or more prior
// States that precede the promise on its world line.  The Value is the

//
// type Promise interface { Atom
