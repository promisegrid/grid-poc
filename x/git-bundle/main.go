package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/packfile"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func createBundle(repoPath, outputFile string, refs []string) error {
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

	// Resolve and write specified references to bundle
	for _, refName := range refs {
		ref, err := repo.Reference(plumbing.ReferenceName(refName), true)
		if err != nil {
			return fmt.Errorf("failed to resolve reference '%s': %w", refName, err)
		}
		if ref.Type() != plumbing.HashReference {
			return fmt.Errorf("reference '%s' is not a hash reference", refName)
		}
		_, err = fmt.Fprintf(f, "%s %s\n", ref.Hash(), ref.Name())
		if err != nil {
			return fmt.Errorf("failed to write reference '%s': %w", refName, err)
		}
	}

	// Write delimiter
	_, err = f.WriteString("\n")
	if err != nil {
		return fmt.Errorf("failed to write delimiter: %w", err)
	}

	// Create packfile encoder
	pw := packfile.NewEncoder(f, repo.Storer, false)

	// Collect object hashes reachable from the specified refs
	var hashes []plumbing.Hash
	seen := make(map[plumbing.Hash]bool)
	for _, refName := range refs {
		ref, err := repo.Reference(plumbing.ReferenceName(refName), true)
		if err != nil {
			return fmt.Errorf("failed to resolve reference '%s': %w", refName, err)
		}

		commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return fmt.Errorf("failed to get commit iterator for reference '%s': %w", refName, err)
		}

		err = commitIter.ForEach(func(c *object.Commit) error {
			if !seen[c.Hash] {
				seen[c.Hash] = true
				hashes = append(hashes, c.Hash)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to iterate commits for reference '%s': %w", refName, err)
		}
	}

	// Encode all reachable objects into the packfile
	_, err = pw.Encode(hashes, 10)
	if err != nil {
		return fmt.Errorf("failed to write packfile: %w", err)
	}

	return nil
}

func main() {
	repoPath := flag.String("repo", "", "Path to the Git repository")
	outputFile := flag.String("out", "repo.bundle", "Output bundle file")
	refs := flag.String("refs", "", "Comma-separated list of references (e.g., refs/heads/main,refs/tags/v1.0)")

	flag.Parse()

	if *repoPath == "" {
		fmt.Println("Error: repository path is required")
		flag.Usage()
		os.Exit(1)
	}

	if *refs == "" {
		fmt.Println("Error: at least one reference is required")
		flag.Usage()
		os.Exit(1)
	}

	refList := parseRefs(*refs)

	err := createBundle(*repoPath, *outputFile, refList)
	if err != nil {
		fmt.Printf("Error creating bundle: %v\n", err)
		return
	}

	fmt.Println("Bundle created successfully!")
}

func parseRefs(refs string) []string {
	var refList []string
	for _, ref := range strings.Split(refs, ",") {
		trimmed := strings.TrimSpace(ref)
		if trimmed != "" {
			refList = append(refList, trimmed)
		}
	}
	return refList
}
