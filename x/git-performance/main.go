package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	. "github.com/stevegt/goadapt"
)

// File represents a file in a git repository
type File struct {
	Name    string
	Content string
}

func main() {

	mode := os.Args[1]

	nStr := os.Args[2]
	N, err := strconv.ParseInt(nStr, 10, 64)
	Ck(err)

	// create a temp directory
	dir, err := ioutil.TempDir("", "example")
	Ck(err)
	// print the temp directory path
	Pl(dir)

	// init a git repository in the temp directory
	repo, err := git.PlainInit(dir, false)
	Ck(err)

	switch mode {
	case "commits":
		// create N commits
		createCommits(repo, dir, N)
		iterateVersions(repo)

	case "branches":
		// create N branches
		createBranches(repo, dir, N)

	case "objects":
		// create N unreferenced objects
		createUnreferencedObjects(repo, dir, N)

	case "referenced":
		// create N referenced objects
		createReferencedObjects(repo, dir, N)

	case "commitWithParents":
		// create a commit with N parents
		createCommitWithParents(repo, dir, N)

	case "repeatability":
		checkRepeateability(repo, dir)

	default:
		Assert(false, "invalid mode %v", mode)
	}

}

// checkRepeateability checks the repeatability of creating a commit
func checkRepeateability(repo *git.Repository, dir string) {
	Pf("Checking repeatability\n%v\n", dir)
	// create a commit
	orig := createCommit(repo, dir, nil, nil)
	now := time.Now()
	// create a new commit
	file := &File{
		Name:    "example.txt",
		Content: Spf("time: %v", now),
	}
	commitOptions := &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@example.com",
			When:  now,
		},
	}
	hash1 := createCommit(repo, dir, file, commitOptions)
	// time.Sleep(10 * time.Second)
	// remove it with a hard reset
	w, err := repo.Worktree()
	Ck(err)
	err = w.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: orig})
	Ck(err)
	// time.Sleep(10 * time.Second)
	// create the same commit
	hash2 := createCommit(repo, dir, file, commitOptions)
	Assert(hash1 == hash2, "hashes do not match: %v %v", hash1, hash2)
	Pf("hashes match: %v %v\n", hash1, hash2)
}

// createCommitWithParents creates a commit with N parents in the git repository
func createCommitWithParents(repo *git.Repository, dir string, N int64) {

	w, err := repo.Worktree()
	Ck(err)

	// create N commits
	parents := make([]plumbing.Hash, N)
	for i := int64(0); i < N; i++ {
		// modify the file in the temp directory
		txt := Spf("time: %v", time.Now())
		err := ioutil.WriteFile(filepath.Join(dir, "example.txt"), []byte(txt), 0644)
		Ck(err)
		// add the file to the git repository
		_, err = w.Add("example.txt")
		Ck(err)
		// commit the file to the git repository
		hash, err := w.Commit("example.txt", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "John Doe",
				Email: "john@example.com",
				When:  time.Now(),
			},
		})
		Ck(err)
		parents[i] = hash
	}

	// create a new commit with N parents
	start := time.Now()
	// modify the file in the temp directory
	txt := Spf("time: %v", time.Now())
	err = ioutil.WriteFile(filepath.Join(dir, "example.txt"), []byte(txt), 0644)
	Ck(err)
	// add the file to the git repository
	_, err = w.Add("example.txt")
	Ck(err)
	// commit the file to the git repository
	msg := Spf("commit with %v parents", N)
	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@example.com",
			When:  time.Now(),
		},
		Parents: parents,
	})
	Ck(err)

	stop := time.Now()
	opsPerSec := float64(N) / stop.Sub(start).Seconds()
	Pf("Commit with parents: %v\n", N)
	Pf("Total time: %v\n", stop.Sub(start))
	Pf("Ops per second: %v\n", opsPerSec)
	Pf("Time per op: %v\n", stop.Sub(start)/time.Duration(N))
}

// createReferencedObjects creates N referenced objects in the git repository
func createReferencedObjects(repo *git.Repository, dir string, N int64) {
	start := time.Now()
	for i := int64(0); i < N; i++ {
		createReferencedObject(repo, dir, i)
	}
	stop := time.Now()
	opsPerSec := float64(N) / stop.Sub(start).Seconds()

	Pl("Referenced objects: ", N)
	Pf("Total time: %v\n", stop.Sub(start))
	Pl("Ops per second: ", opsPerSec)
	Pf("Time per op: %v\n", stop.Sub(start)/time.Duration(N))
}

// createReferencedObject creates a referenced object in the git repository
func createReferencedObject(repo *git.Repository, dir string, i int64) {
	// create a new blob
	blob := repo.Storer.NewEncodedObject()
	blob.SetType(plumbing.BlobObject)
	txt := Spf("blob content %v", i)
	// set the blob content
	w, err := blob.Writer()
	Ck(err)
	w.Write([]byte(txt))
	err = w.Close()
	Ck(err)
	// save the new blob object
	hash, err := repo.Storer.SetEncodedObject(blob)
	Ck(err)

	// create a reference to the new blob object
	ref := plumbing.NewHashReference(plumbing.ReferenceName(Spf("refs/heads/ref-%v", i)), hash)
	err = repo.Storer.SetReference(ref)
	Ck(err)
}

// createUnreferencedObjects creates N unreferenced objects in the git repository
func createUnreferencedObjects(repo *git.Repository, dir string, N int64) {
	start := time.Now()
	for i := int64(0); i < N; i++ {
		createUnreferencedObject(repo, dir, i)
	}
	stop := time.Now()
	opsPerSec := float64(N) / stop.Sub(start).Seconds()

	Pl("Unreferenced objects: ", N)
	Pf("Total time: %v\n", stop.Sub(start))
	Pl("Ops per second: ", opsPerSec)
	Pf("Time per op: %v\n", stop.Sub(start)/time.Duration(N))
}

// createUnreferencedObject creates an unreferenced object in the git repository
func createUnreferencedObject(repo *git.Repository, dir string, i int64) {
	// create a new blob
	blob := repo.Storer.NewEncodedObject()
	blob.SetType(plumbing.BlobObject)
	txt := Spf("blob content %v", i)
	// set the blob content
	w, err := blob.Writer()
	Ck(err)
	w.Write([]byte(txt))
	err = w.Close()
	Ck(err)
	_, err = repo.Storer.SetEncodedObject(blob)
	Ck(err)
}

// createCommits creates N commits in the git repository
func createCommits(repo *git.Repository, dir string, N int64) {
	start := time.Now()
	for i := 0; i < int(N); i++ {
		createCommit(repo, dir, nil, nil)
	}
	stop := time.Now()
	opsPerSec := float64(N) / stop.Sub(start).Seconds()

	Pl("Commits: ", N)
	Pf("Total time: %v\n", stop.Sub(start))
	Pl("Ops per second: ", opsPerSec)
	Pf("Time per op: %v\n", stop.Sub(start)/time.Duration(N))
}

// iterateVersions iterates over all of the versions of example.txt in
// the git repository, retrieving the content of each version
func iterateVersions(repo *git.Repository) {
	// get the HEAD reference
	headRef, err := repo.Head()
	Ck(err)

	// iterate over all of the commits going back to the initial commit
	refHash := headRef.Hash()
	for i := 0; ; i-- {
		// get the commit object
		commit, err := repo.CommitObject(refHash)
		if err != nil {
			break
		}
		// get the file system tree of the commit
		tree, err := commit.Tree()
		Ck(err)
		// get the file system tree entry for example.txt
		file, err := tree.File("example.txt")
		if err != nil {
			// file does not exist in this commit; we're done
			break
		}
		// get the blob of the file system tree entry
		blobReader, err := file.Blob.Reader()
		Ck(err)
		// get the content of the blob
		var content bytes.Buffer
		_, err = content.ReadFrom(blobReader)
		Ck(err)
		Pf("Commit %v: %v\n", i, string(content.Bytes()))

		// get the parent commit
		parentHashes := commit.ParentHashes
		if len(parentHashes) == 0 {
			break
		}
		Assert(len(parentHashes) == 1, "expected 1 parent")
		parentHash := parentHashes[0]
		refHash = parentHash
	}

}

// createBranches creates N branches in the git repository
func createBranches(repo *git.Repository, dir string, N int64) {
	start := time.Now()
	for i := 0; i < int(N); i++ {
		createBranch(repo, dir, i)
	}
	stop := time.Now()
	opsPerSec := float64(N) / stop.Sub(start).Seconds()

	Pl("Branches: ", N)
	Pf("Total time: %v\n", stop.Sub(start))
	Pl("Ops per second: ", opsPerSec)
	Pf("Time per op: %v\n", stop.Sub(start)/time.Duration(N))
}

// createBranch creates a branch in the git repository
func createBranch(repo *git.Repository, dir string, i int) {
	// create a new branch reference from current HEAD
	headRef, err := repo.Head()
	Ck(err)
	branchName := Spf("branch-%v", i)
	branchRef := plumbing.NewHashReference(
		plumbing.ReferenceName("refs/heads/"+branchName),
		plumbing.Hash(headRef.Hash()),
	)
	// save the new branch reference
	err = repo.Storer.SetReference(branchRef)
	Ck(err)
}

// createCommit creates a commit in the git repository
func createCommit(repo *git.Repository, dir string, file *File, commitOptions *git.CommitOptions) plumbing.Hash {

	if file == nil {
		file = &File{
			Name:    "example.txt",
			Content: Spf("time: %v", time.Now()),
		}
	}

	// write to a file in the temp directory
	err := ioutil.WriteFile(filepath.Join(dir, file.Name), []byte(file.Content), 0644)
	Ck(err)

	// add the file to the git repository
	w, err := repo.Worktree()
	Ck(err)
	_, err = w.Add(file.Name)
	Ck(err)

	// set default options if commitOptions is nil
	if commitOptions == nil {
		commitOptions = &git.CommitOptions{
			Author: &object.Signature{
				Name:  "John Doe",
				Email: "john@example.com",
				When:  time.Now(),
			},
		}
	}

	// commit the file to the git repository
	hash, err := w.Commit(file.Name, commitOptions)
	Ck(err)

	return hash
}
