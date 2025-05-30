package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"

	. "github.com/stevegt/goadapt"
)

// CommitData represents the structure of a commit in CBOR format.
type CommitData struct {
	Hash           string            `cbor:"hash"`
	Tree           string            `cbor:"tree"`
	Parents        []string          `cbor:"parents"`
	AuthorName     string            `cbor:"author_name"`
	AuthorEmail    string            `cbor:"author_email"`
	AuthorDate     time.Time         `cbor:"author_date"`
	CommitterName  string            `cbor:"committer_name"`
	CommitterEmail string            `cbor:"committer_email"`
	CommitterDate  time.Time         `cbor:"committer_date"`
	Message        string            `cbor:"message"`
	Trees          []TreeData        `cbor:"trees"`
	Blobs          []BlobData        `cbor:"blobs"`
	Branches       map[string]string `cbor:"branches"`
	Tags           map[string]string `cbor:"tags"`
}

// TreeData represents the structure of a tree object in CBOR format.
type TreeData struct {
	Hash    string      `cbor:"hash"`
	Entries []TreeEntry `cbor:"entries"`
}

// TreeEntry represents an entry within a tree object.
type TreeEntry struct {
	Mode string `cbor:"mode"`
	Name string `cbor:"name"`
	Hash string `cbor:"hash"`
}

// BlobData represents the structure of a blob object in CBOR format.
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
		fmt.Fprintf(os.Stderr, "  cbor2json\n")
		fmt.Fprintf(os.Stderr, "  cbor2dot\n")
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
	case "cbor2json":
		cbor2json()
	case "cbor2dot":
		cbor2dot()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

// git2cbor converts a git commit object to CBOR format, including new and changed tree and blob objects.
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
		Hash:           hash.String(),
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
		Branches:       make(map[string]string),
		Tags:           make(map[string]string),
	}

	for i, parentHash := range commit.ParentHashes {
		commitData.Parents[i] = parentHash.String()
	}

	// Collect branches and tags
	err = collectBranchesAndTags(repo, &commitData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting branches and tags: %v\n", err)
		os.Exit(1)
	}

	// Collect trees from parent commits
	parentTrees, err := collectParentTrees(repo, commit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting parent trees: %v\n", err)
		os.Exit(1)
	}

	// Collect new and changed tree objects
	if err := collectNewAndChangedTrees(repo, commit.TreeHash, &commitData.Trees, parentTrees); err != nil {
		fmt.Fprintf(os.Stderr, "Error collecting tree objects: %v\n", err)
		os.Exit(1)
	}

	// Collect blob objects from the new and changed trees
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

// collectParentTrees collects all tree hashes from parent commits.
func collectParentTrees(repo *git.Repository, commit *object.Commit) (map[string]struct{}, error) {
	parentTrees := make(map[string]struct{})
	for _, parentHash := range commit.ParentHashes {
		parentCommit, err := repo.CommitObject(parentHash)
		if err != nil {
			return nil, fmt.Errorf("error retrieving parent commit '%s': %w", parentHash.String(), err)
		}
		if err := traverseTree(repo, parentCommit.TreeHash, parentTrees); err != nil {
			return nil, fmt.Errorf("error traversing tree for parent commit '%s': %w", parentHash.String(), err)
		}
	}
	return parentTrees, nil
}

// traverseTree recursively traverses a tree and records all tree hashes.
func traverseTree(repo *git.Repository, treeHash plumbing.Hash, treeSet map[string]struct{}) error {
	tree, err := repo.TreeObject(treeHash)
	if err != nil {
		return fmt.Errorf("error retrieving tree object '%s': %w", treeHash.String(), err)
	}

	treeSet[treeHash.String()] = struct{}{}

	for _, entry := range tree.Entries {
		if entry.Mode == filemode.Dir {
			if err := traverseTree(repo, entry.Hash, treeSet); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectNewAndChangedTrees collects trees that are new or have changed compared to parent commits.
func collectNewAndChangedTrees(repo *git.Repository, currentTreeHash plumbing.Hash, trees *[]TreeData, parentTrees map[string]struct{}) error {
	// Check if the current tree is already present in parent trees
	if _, exists := parentTrees[currentTreeHash.String()]; exists {
		// No changes in this tree
		return nil
	}

	// Retrieve the current tree object
	tree, err := repo.TreeObject(currentTreeHash)
	if err != nil {
		return fmt.Errorf("error retrieving tree object '%s': %w", currentTreeHash.String(), err)
	}

	treeData := TreeData{
		Hash:    tree.Hash.String(),
		Entries: []TreeEntry{},
	}

	for _, entry := range tree.Entries {
		treeEntry := TreeEntry{
			Mode: entry.Mode.String(),
			Name: entry.Name,
			Hash: entry.Hash.String(),
		}
		treeData.Entries = append(treeData.Entries, treeEntry)

		if entry.Mode == filemode.Dir {
			// Recursively collect subtrees
			if err := collectNewAndChangedTrees(repo, entry.Hash, trees, parentTrees); err != nil {
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
				Content: getBlobContent(blob),
			}
			// Avoid duplicate blobs
			if !blobExists(*blobs, blobData.Hash) {
				*blobs = append(*blobs, blobData)
			}
		}
	}
	return nil
}

// getBlobContent returns the content of a blob object.
func getBlobContent(blob *object.Blob) (content []byte) {
	reader, err := blob.Reader()
	if err != nil {
		return nil
	}
	defer reader.Close()
	content, err = io.ReadAll(reader)
	if err != nil {
		return nil
	}
	return content
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
	_, err = repo.TreeObject(treeHash)
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

	// Prepare parent hashes
	parentHashes := parentCommitHashes(parentCommits)

	// Create the commit object
	commit := &object.Commit{
		Author:       *author,
		Committer:    *committer,
		Message:      commitData.Message,
		TreeHash:     treeHash,
		ParentHashes: parentHashes,
	}

	// Compute the commit hash to verify
	computedHash, err := recalculateCommitHash(commit)
	Ck(err)

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

	// Recreate branches and tags
	err = recreateBranchesAndTags(repo, commitData.Branches, commitData.Tags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error recreating branches and tags: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully stored commit %s\n", computedHash.String())
}

// writeTreeToRepo writes a tree object to the repository.
func writeTreeToRepo(repo *git.Repository, treeData TreeData) error {
	var tree object.Tree
	for _, entry := range treeData.Entries {
		mode, err := getFileMode(entry.Mode)
		if err != nil {
			return fmt.Errorf("invalid mode '%s' for entry '%s': %w", entry.Mode, entry.Name, err)
		}
		tree.Entries = append(tree.Entries, object.TreeEntry{
			Name: entry.Name,
			Mode: mode,
			Hash: plumbing.NewHash(entry.Hash),
		})
	}

	// Create a new encoded object for the tree
	obj := repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.TreeObject)

	if err := tree.Encode(obj); err != nil {
		return fmt.Errorf("failed to encode tree object: %w", err)
	}

	// Store the tree in the repository
	_, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("failed to store tree object: %w", err)
	}

	return nil
}

// writeBlobToRepo writes a blob object to the repository.
func writeBlobToRepo(repo *git.Repository, blobData BlobData) error {
	// Create a new encoded object for the blob
	obj := repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)

	// Get the writer for the encoded object
	w, err := obj.Writer()
	if err != nil {
		return fmt.Errorf("failed to get writer for blob object: %w", err)
	}

	// Write blob content to the writer
	_, err = w.Write(blobData.Content)
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write blob content: %w", err)
	}

	// Close the writer to finalize the object
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer for blob object: %w", err)
	}

	// Store the blob in the repository
	_, err = repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("failed to store blob object: %w", err)
	}

	return nil
}

// getFileMode converts string mode to filemode.FileMode
func getFileMode(modeStr string) (filemode.FileMode, error) {
	switch modeStr {
	case "0100644":
		return filemode.Regular, nil
	case "0100755":
		return filemode.Executable, nil
	case "0040000":
		return filemode.Dir, nil
	case "0160000":
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

	fmt.Println(diag)
}

// cbor2json converts CBOR data to an indented, pretty-printed JSON representation.
func cbor2json() {
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

	// Marshal CommitData to indented JSON
	jsonData, err := json.MarshalIndent(commitData, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	// Write JSON to stdout
	_, err = os.Stdout.Write(jsonData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON to stdout: %v\n", err)
		os.Exit(1)
	}
}

// cbor2dot converts CBOR data to a Graphviz DOT file representing the structure of the CBOR object.
func cbor2dot() {
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

	// Start building the DOT graph
	fmt.Println("digraph CBOR {")
	fmt.Println("  graph [rankdir=LR];")
	fmt.Println("  node [fontname=\"Helvetica\"];")

	// Short hash helper
	shortHash := func(fullHash string) string {
		if len(fullHash) >= 8 {
			return fullHash[:8]
		}
		return fullHash
	}

	// Extract the first line of the commit message
	firstLine := commitData.Message
	if idx := strings.Index(firstLine, "\n"); idx != -1 {
		firstLine = firstLine[:idx]
	}
	escapedMsg := escapeDotString(firstLine)

	// Central node for commit message
	commitMsgNode := fmt.Sprintf("commit_msg_%s", shortHash(commitData.Hash))
	fmt.Printf("  %s [label=\"%s\", shape=oval, style=filled, color=lightblue];\n", commitMsgNode, escapedMsg)

	// Node for commit hash
	commitHashNode := fmt.Sprintf("hash_%s", shortHash(commitData.Hash))
	fmt.Printf("  %s [label=\"%s\", shape=rectangle, style=filled, color=lightgray];\n", commitHashNode, shortHash(commitData.Hash))
	fmt.Printf("  %s -> %s;\n", commitMsgNode, commitHashNode)

	// Parent commits
	for _, parentHash := range commitData.Parents {
		parentShortHash := shortHash(parentHash)
		parentNodeID := fmt.Sprintf("parent_%s", parentShortHash)
		fmt.Printf("  %s [label=\"%s\", shape=rectangle, style=filled, color=lightgray];\n", parentNodeID, parentShortHash)
		fmt.Printf("  %s -> %s;\n", parentNodeID, commitMsgNode)
	}

	// Branches
	for branchName, branchHash := range commitData.Branches {
		branchNodeID := fmt.Sprintf("branch_%s", escapeDotString(branchName))
		branchLabel := fmt.Sprintf("Branch: %s\\n%s", escapeDotString(branchName), shortHash(string(branchHash)))
		fmt.Printf("  %s [label=\"%s\", shape=ellipse, style=filled, color=lightcoral];\n", branchNodeID, branchLabel)
		fmt.Printf("  %s -> %s;\n", branchNodeID, commitMsgNode)
	}

	// Tags
	for tagName, tagHash := range commitData.Tags {
		tagNodeID := fmt.Sprintf("tag_%s", escapeDotString(tagName))
		tagLabel := fmt.Sprintf("Tag: %s\\n%s", escapeDotString(tagName), shortHash(string(tagHash)))
		fmt.Printf("  %s [label=\"%s\", shape=diamond, style=filled, color=gold];\n", tagNodeID, tagLabel)
		fmt.Printf("  %s -> %s;\n", tagNodeID, commitMsgNode)
	}

	// Trees
	for _, tree := range commitData.Trees {
		treeShortHash := shortHash(tree.Hash)
		treeNodeID := fmt.Sprintf("tree_%s", treeShortHash)
		fmt.Printf("  %s [label=\"Tree: %s\", shape=folder, style=filled, color=lightgreen];\n", treeNodeID, treeShortHash)
		// point tree nodes to commitMsgNode
		fmt.Printf("  %s -> %s;\n", treeNodeID, commitMsgNode)

		for _, entry := range tree.Entries {
			if entry.Mode == "040000" { // Directory
				subTreeShortHash := shortHash(entry.Hash)
				subTreeNodeID := fmt.Sprintf("tree_%s", subTreeShortHash)
				fmt.Printf("  %s [label=\"Tree: %s\", shape=folder, style=filled, color=lightgreen];\n", subTreeNodeID, subTreeShortHash)
				fmt.Printf("  %s -> %s;\n", subTreeNodeID, treeNodeID)
			} else { // Blob or file
				blobShortHash := shortHash(entry.Hash)
				blobNodeID := fmt.Sprintf("blob_%s", blobShortHash)
				fmt.Printf("  %s [label=\"Blob: %s (%s)\", shape=note, style=filled, color=yellow];\n", blobNodeID, escapeDotString(entry.Name), blobShortHash)
				fmt.Printf("  %s -> %s;\n", blobNodeID, treeNodeID)
			}
		}
	}

	// Blobs
	for _, blob := range commitData.Blobs {
		blobShortHash := shortHash(blob.Hash)
		blobNodeID := fmt.Sprintf("blob_content_%s", blobShortHash)
		contentSnippet := string(blob.Content)
		if len(contentSnippet) > 20 {
			contentSnippet = contentSnippet[:20] + "..."
		}
		escapedContent := escapeDotString(contentSnippet)
		fmt.Printf("  %s [label=\"Blob Content: %s\", shape=note, style=filled, color=yellow];\n", blobNodeID, escapedContent)
		fmt.Printf("  %s -> blob_%s;\n", blobNodeID, blobShortHash)
	}

	fmt.Println("}")
}

// escapeDotString escapes special characters in strings for DOT format.
func escapeDotString(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
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
	// Create a new encoded object for the commit
	obj := repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.CommitObject)

	// Encode the commit into the writer
	if err := commit.Encode(obj); err != nil {
		return fmt.Errorf("failed to encode commit object: %w", err)
	}

	// Store the commit in the repository
	_, err := repo.Storer.SetEncodedObject(obj)
	if err != nil {
		return fmt.Errorf("failed to store commit object: %w", err)
	}

	return nil
}

// recalculateCommitHash recalculates the commit hash based on its content.
func recalculateCommitHash(commit *object.Commit) (plumbing.Hash, error) {
	mo := plumbing.MemoryObject{}
	mo.SetType(plumbing.CommitObject)
	err := commit.Encode(&mo)
	if err != nil {
		return plumbing.ZeroHash, err
	}
	return mo.Hash(), nil
}

// collectBranchesAndTags collects all branches and tags in the repository and adds them to commitData.
func collectBranchesAndTags(repo *git.Repository, commitData *CommitData) error {
	branches, err := repo.Branches()
	if err != nil {
		return fmt.Errorf("error retrieving branches: %w", err)
	}

	err = branches.ForEach(func(b *plumbing.Reference) error {
		branchName := b.Name().Short()
		commitHash := b.Hash().String()
		commitData.Branches[branchName] = commitHash
		return nil
	})
	if err != nil {
		return fmt.Errorf("error iterating branches: %w", err)
	}

	tags, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("error retrieving tags: %w", err)
	}

	err = tags.ForEach(func(t *plumbing.Reference) error {
		tagName := t.Name().Short()
		target, err := repo.Tag(tagName)
		if err != nil {
			// It might not be an annotated tag; try to resolve to commit
			commit, err := repo.CommitObject(t.Hash())
			if err != nil {
				return fmt.Errorf("error resolving tag '%s': %w", tagName, err)
			}
			commitData.Tags[tagName] = commit.Hash.String()
			return nil
		}
		commitData.Tags[tagName] = string(target.Target())
		return nil
	})
	if err != nil {
		return fmt.Errorf("error iterating tags: %w", err)
	}

	return nil
}

// recreateBranchesAndTags recreates branches and tags in the repository from the provided maps.
func recreateBranchesAndTags(repo *git.Repository, branches map[string]string, tags map[string]string) error {
	// Recreate branches
	for branchName, commitHash := range branches {
		refName := plumbing.NewBranchReferenceName(branchName)
		hash := plumbing.NewHash(commitHash)
		ref := plumbing.NewHashReference(refName, hash)
		err := repo.Storer.SetReference(ref)
		if err != nil {
			return fmt.Errorf("error recreating branch '%s': %w", branchName, err)
		}
	}

	// Recreate tags
	for tagName, targetHash := range tags {
		refName := plumbing.NewTagReferenceName(tagName)
		hash := plumbing.NewHash(targetHash)
		ref := plumbing.NewHashReference(refName, hash)
		err := repo.Storer.SetReference(ref)
		if err != nil {
			return fmt.Errorf("error recreating tag '%s': %w", tagName, err)
		}
	}

	return nil
}
