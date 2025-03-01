package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

// StoreDataInIPFS stores a given data string in IPFS and returns its content hash.
func StoreDataInIPFS(data string) (string, error) {
	sh := shell.NewShell("localhost:5001")
	cid, err := sh.Add(strings.NewReader(data))
	if err != nil {
		return "", err
	}
	return cid, nil
}

// RetrieveDataFromIPFS fetches data from IPFS based on its content hash.
func RetrieveDataFromIPFS(cid string) (string, error) {
	sh := shell.NewShell("localhost:5001")
	reader, err := sh.Cat(cid)
	if err != nil {
		return "", err
	}
	buf := new(strings.Builder)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func main() {
	data := "PromiseGrid DAG Node Example Data"
	cid, err := StoreDataInIPFS(data)
	if err != nil {
		log.Fatalf("Error storing data in IPFS: %v", err)
	}
	fmt.Printf("Data stored in IPFS with CID: %s\n", cid)
	retrieved, err := RetrieveDataFromIPFS(cid)
	if err != nil {
		log.Fatalf("Error retrieving data: %v", err)
	}
	fmt.Printf("Retrieved Data: %s\n", retrieved)
}
