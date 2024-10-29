package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <subcommand> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Subcommands:\n")
		fmt.Fprintf(os.Stderr, "  git2cbor <ref>\n")
		fmt.Fprintf(os.Stderr, "  cbor2git\n")
		fmt.Fprintf(os.Stderr, "  cbor2diag\n")
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "git2cbor":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: %s git2cbor <ref>\n", os.Args[0])
			os.Exit(1)
		}
		git2cbor(os.Args[2])
	case "cbor2git":
		cbor2git()
	case "cbor2diag":
		cbor2diag()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

// git2cbor converts a git commit object to CBOR format.
func git2cbor(ref string) {
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
	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving ref '%s': %v\n", ref, err)
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

// cbor2git converts CBOR data to a git commit object.
func cbor2git() {
	// Read CBOR data from stdin
	cborData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading CBOR data: %v\n", err)
		os.Exit(1)
	}

	// Decode CBOR data into CommitData
	var commitData CommitData
	err = cbor.Unmarshal(cborData, &commitData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding CBOR data: %v\n", err)
		os.Exit(1)
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

	// Prepare parent commits
	var parentCommits []*object.Commit
	for _, ph := range commitData.Parents {
		pHash, err := plumbing.NewHash(ph)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid parent hash '%s': %v\n", ph, err)
			os.Exit(1)
		}
		parentCommit, err := repo.CommitObject(*pHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving parent commit '%s': %v\n", ph, err)
			os.Exit(1)
		}
		parentCommits = append(parentCommits, parentCommit)
	}

	// Prepare the tree
	treeHash, err := plumbing.NewHash(commitData.Tree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid tree hash '%s': %v\n", commitData.Tree, err)
		os.Exit(1)
	}
	tree, err := repo.TreeObject(*treeHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving tree object: %v\n", err)
		os.Exit(1)
	}

	// Create author and committer signatures
	author := &object.Signature{
		Name:  commitData.AuthorName,
		Email: commitData.AuthorEmail,
		When:  commitData.AuthorDate,
	}
	committer := &object.Signature{
		Name:  commitData.CommitterName,
		Email: commitData.CommitterEmail,
		When:  commitData.CommitterDate,
	}

	// Create the commit object
	commit := &object.Commit{
		Author:       *author,
		Committer:    *committer,
		Message:      commitData.Message,
		TreeHash:     *treeHash,
		ParentHashes: parentCommitsHashes(parentCommits),
	}

	// Serialize the commit to get the correct hash
	commitHash, err := commit.Hash()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing commit hash: %v\n", err)
		os.Exit(1)
	}

	// Verify that the computed hash matches the provided hash
	if commitHash.String() != commitData.Hash {
		fmt.Fprintf(os.Stderr, "Computed hash '%s' does not match provided hash '%s'\n", commitHash.String(), commitData.Hash)
		os.Exit(1)
	}

	// Write the commit object to the repository
	err = writeCommitToRepo(repo, commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing commit to repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully stored commit %s\n", commitHash.String())
}

// cbor2diag emits a human-readable CBOR diagnostic representation of the CBOR object.
func cbor2diag() {
	// Read CBOR data from stdin
	cborData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading CBOR data: %v\n", err)
		os.Exit(1)
	}

	// Convert CBOR to diagnostic format
	diag, err := cbor.Diagnostic(cborData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating CBOR diagnostic: %v\n", err)
		os.Exit(1)
	}

	// Print the diagnostic representation
	fmt.Println(diag)
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

// parentCommitsHashes extracts the hashes from parent commit objects.
func parentCommitsHashes(parents []*object.Commit) []plumbing.Hash {
	hashes := make([]plumbing.Hash, len(parents))
	for i, parent := range parents {
		hashes[i] = parent.Hash
	}
	return hashes
}

// writeCommitToRepo writes the commit object to the repository's object store.
func writeCommitToRepo(repo *git.Repository, commit *object.Commit) error {
	objWriter, err := repo.Storer.NewEncodedObject()
	if err != nil {
		return fmt.Errorf("failed to create new encoded object: %w", err)
	}

	if err := commit.Encode(objWriter); err != nil {
		return fmt.Errorf("failed to encode commit object: %w", err)
	}

	commitHash, err := repo.Storer.SetEncodedObject(objWriter)
	if err != nil {
		return fmt.Errorf("failed to store commit object: %w", err)
	}

	// Optionally, update references as needed here

	// Verify that the stored hash matches
	if commitHash.String() != commit.Hash.String() {
		return fmt.Errorf("stored commit hash '%s' does not match expected hash '%s'", commitHash.String(), commit.Hash.String())
	}

	return nil
}
