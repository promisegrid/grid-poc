package main

import (
	"fmt"
	// "io"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func createBundle(repoPath, outputFile string) error {
	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Create the bundle file
	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create bundle file: %w", err)
	}
	defer f.Close()

	// Write bundle header
	_, err = f.WriteString("# v3 git bundle\n")
	if err != nil {
		return fmt.Errorf("failed to write bundle header: %w", err)
	}

	// Get all references
	refs, err := repo.References()
	if err != nil {
		return fmt.Errorf("failed to get references: %w", err)
	}

	// Write references to bundle
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			_, err := fmt.Fprintf(f, "%s %s\n", ref.Hash(), ref.Name())
			if err != nil {
				return fmt.Errorf("failed to write reference: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to write references: %w", err)
	}

	// Write delimiter
	_, err = f.WriteString("\n")
	if err != nil {
		return fmt.Errorf("failed to write delimiter: %w", err)
	}

	// Create packfile encoder
	pw := packfile.NewEncoder(f, repo.Storer, false)

	// Get all object hashes
	objectIter, err := repo.Objects()
	if err != nil {
		return fmt.Errorf("failed to get objects: %w", err)
	}

	var hashes []plumbing.Hash
	err = objectIter.ForEach(func(obj object.Object) error {
		hashes = append(hashes, obj.ID())
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to collect object hashes: %w", err)
	}

	// Encode all objects into the packfile
	_, err = pw.Encode(hashes, 10)
	if err != nil {
		return fmt.Errorf("failed to write packfile: %w", err)
	}

	return nil
}

func main() {
	repoPath := "/path/to/your/repo"
	outputFile := "repo.bundle"

	err := createBundle(repoPath, outputFile)
	if err != nil {
		fmt.Printf("Error creating bundle: %v\n", err)
		return
	}

	fmt.Println("Bundle created successfully!")
}
