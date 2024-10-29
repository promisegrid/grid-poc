package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/fxamacker/cbor/v2"
)

type CommitData struct {
	Hash           string    `cbor:"hash"`
	Tree           string    `cbor:"tree"`
	Parents        []string  `cbor:"parents"`
	AuthorName     string    `cbor:"author_name"`
	AuthorEmail    string    `cbor:"author_email"`
	AuthorDate     time.Time `cbor:"author_date"`
	CommitterName  string    `cbor:"committer_name"`
	CommitterEmail string    `cbor:"committer_email"`
	CommitterDate  time.Time `cbor:"committer_date"`
	Message        string    `cbor:"message"`
}

func main() {
	// Determine the reference from command line arguments or default to HEAD
	refName := "HEAD"
	if len(os.Args) > 1 {
		refName = os.Args[1]
	}

	// Find the Git repository directory by walking up the directory tree
	repoPath, err := findGitRepo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding .git directory: %v\n", err)
		os.Exit(1)
	}

	// Open the Git repository
	repo, err := git.PlainOpen(repoPath)
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
		Hash:           commit.Hash.String(),
		Tree:           commit.TreeHash.String(),
		Parents:        make([]string, len(commit.ParentHashes)),
		AuthorName:     commit.Author.Name,
		AuthorEmail:    commit.Author.Email,
		AuthorDate:     commit.Author.When,
		CommitterName:  commit.Committer.Name,
		CommitterEmail: commit.Committer.Email,
		CommitterDate:  commit.Committer.When,
		Message:        commit.Message,
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

// findGitRepo walks up the directory tree from the current directory to find a directory containing a .git folder.
// It returns the path to the repository root directory, or an error if not found.
func findGitRepo() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current working directory: %w", err)
	}

	for {
		gitPath := filepath.Join(currentDir, ".git")
		info, err := os.Stat(gitPath)
		if err == nil && info.IsDir() {
			return currentDir, nil
		}

		// Move up to the parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the root directory
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf(".git directory not found in any parent directories")
}
