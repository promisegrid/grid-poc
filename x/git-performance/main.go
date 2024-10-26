package main

import (
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

func main() {

	nStr := os.Args[1]
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

	// create N commits
	// createCommits(repo, dir, N)

	// create N branches
	// createBranches(repo, dir, N)

	// create N unreferenced objects
	// createUnreferencedObjects(repo, dir, N)

	// create N referenced objects
	// createReferencedObjects(repo, dir, N)

	// create a commit with N parents
	createCommitWithParents(repo, dir, N)
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
func createCommits(repo *git.Repository, dir string, N int) {
	start := time.Now()
	for i := 0; i < N; i++ {
		createCommit(repo, dir)
	}
	stop := time.Now()
	opsPerSec := float64(N) / stop.Sub(start).Seconds()

	Pl("Commits: ", N)
	Pf("Total time: %v\n", stop.Sub(start))
	Pl("Ops per second: ", opsPerSec)
	Pf("Time per op: %v\n", stop.Sub(start)/time.Duration(N))
}

// createBranches creates N branches in the git repository
func createBranches(repo *git.Repository, dir string, N int) {
	start := time.Now()
	for i := 0; i < N; i++ {
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
func createCommit(repo *git.Repository, dir string) {

	// write to a file in the temp directory
	txt := Spf("time: %v", time.Now())
	err := ioutil.WriteFile(filepath.Join(dir, "example.txt"), []byte(txt), 0644)
	Ck(err)

	// add the file to the git repository
	w, err := repo.Worktree()
	Ck(err)
	_, err = w.Add("example.txt")
	Ck(err)

	// commit the file to the git repository
	_, err = w.Commit("example.txt", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@example.com",
			When:  time.Now(),
		},
	})
}
