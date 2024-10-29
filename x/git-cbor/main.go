package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	cborData, err := readAll(os.Stdin)
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

	// Prepare parent hashes
	var parentPlumbingHashes []plumbing.Hash
	for _, ph := range commitData.Parents {
		h, err := plumbing.NewHash(ph)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid parent hash '%s': %v\n", ph, err)
			os.Exit(1)
		}
		parentPlumbingHashes = append(parentPlumbingHashes, h)
	}

	// Prepare tree hash
	treeHash, err := plumbing.NewHash(commitData.Tree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid tree hash '%s': %v\n", commitData.Tree, err)
		os.Exit(1)
	}

	// Create the commit object
	commit := &objectCommit{
		Hash: plumbing.NewHash(commitData.Hash),
		Author: objectSignature{
			Name:  commitData.AuthorName,
			Email: commitData.AuthorEmail,
			When:  commitData.AuthorDate,
		},
		Committer: objectSignature{
			Name:  commitData.CommitterName,
			Email: commitData.CommitterEmail,
			When:  commitData.CommitterDate,
		},
		Message:      commitData.Message,
		TreeHash:     treeHash,
		ParentHashes: parentPlumbingHashes,
	}

	// XXX should we be using some sort of commit object from the
	// go-git library instead of serializing it ourselves?

	// Serialize the commit object to the correct format
	commitContent, err := serializeCommit(commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing commit: %v\n", err)
		os.Exit(1)
	}

	// Compute the object hash
	objHash := plumbing.ComputeHash(plumbing.CommitObject, commitContent)

	// Verify that the computed hash matches the provided hash
	if objHash.String() != commitData.Hash {
		fmt.Fprintf(os.Stderr, "Computed hash '%s' does not match provided hash '%s'\n", objHash.String(), commitData.Hash)
		os.Exit(1)
	}

	// Store the commit object in the repository
	err = storeObject(repo, objHash, commitContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error storing commit object: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully stored commit %s\n", objHash.String())
}

// cbor2diag emits a human-readable CBOR diagnostic representation of the CBOR object.
func cbor2diag() {
	// Read CBOR data from stdin
	cborData, err := readAll(os.Stdin)
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

// readAll reads all data from the given file.
func readAll(file *os.File) ([]byte, error) {
	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buffer.Write(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// serializeCommit serializes the commit object into the Git commit format.
// XXX is there a better way to do this with the go-git library?
func serializeCommit(commit *objectCommit) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("tree %s\n", commit.TreeHash))
	for _, parent := range commit.ParentHashes {
		buffer.WriteString(fmt.Sprintf("parent %s\n", parent))
	}
	buffer.WriteString(fmt.Sprintf("author %s <%s> %d %s\n",
		commit.Author.Name,
		commit.Author.Email,
		commit.Author.When.Unix(),
		formatTimeZone(commit.Author.When)),
	)
	buffer.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n\n",
		commit.Committer.Name,
		commit.Committer.Email,
		commit.Committer.When.Unix(),
		formatTimeZone(commit.Committer.When)),
	)
	buffer.WriteString(commit.Message)

	return buffer.Bytes(), nil
}

// formatTimeZone formats the timezone offset.
func formatTimeZone(t time.Time) string {
	_, offset := t.Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / 3600
	minutes := (offset % 3600) / 60
	return fmt.Sprintf("%s%02d%02d", sign, hours, minutes)
}

// storeObject writes the object to the Git repository's object store.
// XXX is there a better way to do this with the go-git library?
func storeObject(repo *git.Repository, hash plumbing.Hash, content []byte) error {
	objectDir := filepath.Join(repo.Storer.(*git.MemoryObjectStorer).Root(), "objects", hash.String()[:2])
	objectPath := filepath.Join(objectDir, hash.String()[2:])

	// Create the object directory if it doesn't exist
	err := os.MkdirAll(objectDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create object directory: %w", err)
	}

	// Write the object if it doesn't already exist
	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		// Compress the content using zlib
		var buf bytes.Buffer
		writer := zlib.NewWriter(&buf)
		_, err := writer.Write(content)
		if err != nil {
			return fmt.Errorf("failed to compress object: %w", err)
		}
		writer.Close()

		// Write to file
		err = os.WriteFile(objectPath, buf.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("failed to write object file: %w", err)
		}
	}

	return nil
}

// objectCommit represents a simplified Git commit object.
// XXX use the go-git library's struct
type objectCommit struct {
	Hash         plumbing.Hash
	Author       objectSignature
	Committer    objectSignature
	Message      string
	TreeHash     plumbing.Hash
	ParentHashes []plumbing.Hash
}

// objectSignature represents the author or committer signature in a commit.
// XXX use the go-git library's struct
type objectSignature struct {
	Name  string
	Email string
	When  time.Time
}
