package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/fxamacker/cbor/v2"
)

type CommitData struct {
	Hash    string   `cbor:"hash"`
	Author  string   `cbor:"author"`
	Message string   `cbor:"message"`
	Parents []string `cbor:"parents"`
	Tree    string   `cbor:"tree"`
}

func main() {
	// Determine the reference from command line arguments or default to HEAD
	refName := "HEAD"
	if len(os.Args) > 1 {
		refName = os.Args[1]
	}

	// Open the Git repository in the current directory
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening repository: %v\n", err)
		os.Exit(1)
	}

	// Resolve the reference to a commit hash
	hash, err := repo.ResolveRevision(plumbing.Revision(refName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving ref '%s': %v\n", refName, err)
		os.Exit(1)
	}

	// Get the commit object from the hash
	commit, err := repo.CommitObject(*hash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving commit object: %v\n", err)
		os.Exit(1)
	}

	// Populate the CommitData struct
	commitData := CommitData{
		Hash:    commit.Hash.String(),
		Author:  commit.Author.String(),
		Message: commit.Message,
		Parents: make([]string, len(commit.ParentHashes)),
		Tree:    commit.TreeHash.String(),
	}

	for i, parentHash := range commit.ParentHashes {
		commitData.Parents[i] = parentHash.String()
	}

	// Encode the CommitData to CBOR
	cborBytes, err := cbor.Marshal(commitData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding commit data to CBOR: %v\n", err)
		os.Exit(1)
	}

	// Write the CBOR data to stdout
	_, err = os.Stdout.Write(cborBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CBOR to stdout: %v\n", err)
		os.Exit(1)
	}
}
