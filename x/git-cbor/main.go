package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/fxamacker/cbor/v2"
)

type GitObject struct {
	Type    string `cbor:"type"`
	Content string `cbor:"content"`
}

func main() {
	var ref string
	if len(os.Args) > 1 {
		ref = os.Args[1]
	} else {
		ref = "HEAD"
	}

	// Get the type of the git object
	typeCmd := exec.Command("git", "cat-file", "-t", ref)
	typeOut, err := typeCmd.Output()
	if err != nil {
		log.Fatalf("Error retrieving type for ref '%s': %v", ref, err)
	}
	objType := stringTrimNewline(string(typeOut))

	// Get the content of the git object
	contentCmd := exec.Command("git", "cat-file", "-p", ref)
	contentOut, err := contentCmd.Output()
	if err != nil {
		log.Fatalf("Error retrieving content for ref '%s': %v", ref, err)
	}
	content := string(contentOut)

	// Create GitObject instance
	gitObj := GitObject{
		Type:    objType,
		Content: content,
	}

	// CBOR encode the GitObject
	encoded, err := cbor.Marshal(gitObj)
	if err != nil {
		log.Fatalf("Error encoding GitObject to CBOR: %v", err)
	}

	// Write CBOR-encoded data to stdout
	_, err = os.Stdout.Write(encoded)
	if err != nil {
		log.Fatalf("Error writing CBOR data to stdout: %v", err)
	}
}

// Helper function to trim the newline character from a string
func stringTrimNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
