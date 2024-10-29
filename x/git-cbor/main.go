package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitData struct {
	Hash           string     `cbor:"hash"`
	Tree           string     `cbor:"tree"`
	Parents        []string   `cbor:"parents"`
	AuthorName     string     `cbor:"author_name"`
	AuthorEmail    string     `cbor:"author_email"`
	AuthorDate     time.Time  `cbor:"author_date"`
	CommitterName  string     `cbor:"committer_name"`
	CommitterEmail string     `cbor:"committer_email"`
	CommitterDate  time.Time  `cbor:"committer_date"`
	Message        string     `cbor:"message"`
	Trees          []TreeData `cbor:"trees"`
	Blobs          []BlobData `cbor:"blobs"`
}

type TreeData struct {
	Hash    string      `cbor:"hash"`
	Entries []TreeEntry `cbor:"entries"`
}

type TreeEntry struct {
	Mode string `cbor:"mode"`
	Name string `cbor:"name"`
	Type string `cbor:"type"`
	Hash string `cbor:"hash"`
}

type BlobData struct {
	Hash    string `cbor:"hash"`
	Content []byte `cbor:"content"`
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

// git2cbor converts a git commit object to CBOR format, including tree and blob objects.
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
		Trees:          []TreeData{},
		Blobs:          []BlobData{},
	}

	for i, parentHash := range commit.ParentHashes {
		commitData.Parents[i] = parentHash.String()
	}

	// Collect tree objects starting from the commit's tree
	if err := collectTrees(repo, commit.TreeHash, &commitData.Trees); err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting tree objects: %v\n", err)
		os.Exit(1)
	}

	// Collect blob objects from the trees
	if err := collectBlobs(repo, commit.TreeHash, &commitData.Blobs); err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting blob objects: %v\n", err)
		os.Exit(1)
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

// collectTrees recursively collects tree objects and appends them to trees slice.
func collectTrees(repo *git.Repository, treeHash plumbing.Hash, trees *[]TreeData) error {
	// Check if the tree is already collected
	for _, t := range *trees {
		if t.Hash == treeHash.String() {
			return nil
		}
	}

	tree, err := repo.TreeObject(treeHash)
	if err != nil {
		return fmt.Errorf("error retrieving tree object '%s': %w", treeHash.String(), err)
	}

	treeData := TreeData{
		Hash:    tree.Hash.String(),
		Entries: []TreeEntry{},
	}

	for _, entry := range tree.Entries {
		treeEntry := TreeEntry{
			Mode: entry.Mode.String(),
			Name: entry.Name,
			Type: entry.Mode.Type().String(),
			Hash: entry.Hash.String(),
		}
		treeData.Entries = append(treeData.Entries, treeEntry)

		if entry.Mode == filemode.Dir {
			// Recursively collect subtrees
			if err := collectTrees(repo, entry.Hash, trees); err != nil {
				return err
			}
		}
	}

	*trees = append(*trees, treeData)
	return nil
}

// collectBlobs recursively collects blob objects and appends them to blobs slice.
func collectBlobs(repo *git.Repository, treeHash plumbing.Hash, blobs *[]BlobData) error {
	tree, err := repo.TreeObject(treeHash)
	if err != nil {
		return fmt.Errorf("error retrieving tree object '%s': %w", treeHash.String(), err)
	}

	for _, entry := range tree.Entries {
		if entry.Mode == filemode.Dir {
			// Recursively collect blobs from subtrees
			if err := collectBlobs(repo, entry.Hash, blobs); err != nil {
				return err
			}
		} else if entry.Mode.IsFile() || entry.Mode == filemode.Symlink {
			blob, err := repo.BlobObject(entry.Hash)
			if err != nil {
				return fmt.Errorf("error retrieving blob object '%s': %w", entry.Hash.String(), err)
			}
			blobData := BlobData{
				Hash:    blob.Hash.String(),
				Content: blob.Contents,
			}
			// Avoid duplicate blobs
			if !blobExists(*blobs, blobData.Hash) {
				*blobs = append(*blobs, blobData)
			}
		}
	}
	return nil
}

// blobExists checks if a blob with the given hash already exists in the blobs slice.
func blobExists(blobs []BlobData, hash string) bool {
	for _, b := range blobs {
		if b.Hash == hash {
			return true
		}
	}
	return false
}

// cbor2git converts CBOR data to a git commit object, including tree and blob objects.
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

	// Write blob objects to the repository
	for _, blobData := range commitData.Blobs {
		if err := writeBlobToRepo(repo, blobData); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing blob '%s' to repository: %v\n", blobData.Hash, err)
			os.Exit(1)
		}
	}

	// Write tree objects to the repository
	for _, treeData := range commitData.Trees {
		if err := writeTreeToRepo(repo, treeData); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing tree '%s' to repository: %v\n", treeData.Hash, err)
			os.Exit(1)
		}
	}

	// Prepare parent commits
	var parentCommits []*object.Commit
	for _, ph := range commitData.Parents {
		pHash := plumbing.NewHash(ph)
		parentCommit, err := repo.CommitObject(pHash)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving parent commit '%s': %v\n", ph, err)
			os.Exit(1)
		}
		parentCommits = append(parentCommits, parentCommit)
	}

	// Prepare the tree
	treeHash := plumbing.NewHash(commitData.Tree)
	tree, err := repo.TreeObject(treeHash)
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
		Hash:         plumbing.NewHash(commitData.Hash),
		Author:       *author,
		Committer:    *committer,
		Message:      commitData.Message,
		TreeHash:     treeHash,
		ParentHashes: parentCommitHashes(parentCommits),
	}

	// Compute the commit hash to verify
	computedHash := commit.ID()

	// Verify that the computed hash matches the provided hash
	if computedHash.String() != commitData.Hash {
		fmt.Fprintf(os.Stderr, "Computed hash '%s' does not match provided hash '%s'\n", computedHash.String(), commitData.Hash)
		os.Exit(1)
	}

	// Write the commit object to the repository
	err = writeCommitToRepo(repo, commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing commit to repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully stored commit %s\n", computedHash.String())
}

// writeTreeToRepo writes a tree object to the repository.
func writeTreeToRepo(repo *git.Repository, treeData TreeData) error {
	treeBuilder := object.NewTreeBuilder(repo.Storer)

	for _, entry := range treeData.Entries {
		mode, err := getFileMode(entry.Mode)
		if err != nil {
			return fmt.Errorf("invalid mode '%s' for entry '%s': %w", entry.Mode, entry.Name, err)
		}
		hash := plumbing.NewHash(entry.Hash)
		err = treeBuilder.Insert(entry.Name, hash, mode)
		if err != nil {
			return fmt.Errorf("error inserting entry '%s' into tree builder: %w", entry.Name, err)
		}
	}

	tree, err := treeBuilder.Commit()
	if err != nil {
		return fmt.Errorf("error committing tree builder: %w", err)
	}

	// Ensure the tree hash matches
	if tree.String() != treeData.Hash {
		return fmt.Errorf("computed tree hash '%s' does not match expected hash '%s'", tree.String(), treeData.Hash)
	}

	// Store the tree in the repository
	_, err = repo.Storer.SetEncodedObject(tree)
	if err != nil {
		return fmt.Errorf("error storing tree object: %w", err)
	}

	return nil
}

// writeBlobToRepo writes a blob object to the repository.
func writeBlobToRepo(repo *git.Repository, blobData BlobData) error {
	blob := &object.Blob{
		Hash:        plumbing.NewHash(blobData.Hash),
		Size:        int64(len(blobData.Content)),
		BlobContent: blobData.Content,
	}
	// Encode the blob
	objWriter := repo.Storer.NewEncodedObject()
	if err := blob.Encode(objWriter); err != nil {
		return fmt.Errorf("failed to encode blob object: %w", err)
	}

	// Store the blob in the repository
	storedHash, err := repo.Storer.SetEncodedObject(objWriter)
	if err != nil {
		return fmt.Errorf("failed to store blob object: %w", err)
	}

	// Verify that the stored hash matches
	if storedHash.String() != blobData.Hash {
		return fmt.Errorf("stored blob hash '%s' does not match expected hash '%s'", storedHash.String(), blobData.Hash)
	}

	return nil
}

// getFileMode converts string mode to plumbing.FileMode
func getFileMode(modeStr string) (plumbing.FileMode, error) {
	switch modeStr {
	case "100644":
		return filemode.Regular, nil
	case "100755":
		return filemode.Executable, nil
	case "040000":
		return filemode.Dir, nil
	case "160000":
		return filemode.Submodule, nil
	default:
		return 0, fmt.Errorf("unsupported file mode: %s", modeStr)
	}
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
	diag, err := cbor.Diagnose(cborData)
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

// parentCommitHashes extracts the hashes from parent commit objects.
func parentCommitHashes(parents []*object.Commit) []plumbing.Hash {
	hashes := make([]plumbing.Hash, len(parents))
	for i, parent := range parents {
		hashes[i] = parent.Hash
	}
	return hashes
}

// writeCommitToRepo writes the commit object to the repository's object store.
func writeCommitToRepo(repo *git.Repository, commit *object.Commit) error {
	objWriter := repo.Storer.NewEncodedObject()

	if err := commit.Encode(objWriter); err != nil {
		return fmt.Errorf("failed to encode commit object: %w", err)
	}

	commitHash, err := repo.Storer.SetEncodedObject(objWriter)
	if err != nil {
		return fmt.Errorf("failed to store commit object: %w", err)
	}

	// Verify that the stored hash matches
	if commitHash.String() != commit.Hash.String() {
		return fmt.Errorf("stored commit hash '%s' does not match expected hash '%s'", commitHash.String(), commit.Hash.String())
	}

	return nil
}
