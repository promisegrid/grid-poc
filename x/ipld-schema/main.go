package main

import (
	"fmt"
	"os"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/node/bindnode"
)

// Define Go structures that match our schema
type Person struct {
	Name    string
	Age     int
	Friends []string
}

func main() {
	// Define an IPLD Schema as a string
	schema := `type Person struct {
		name String
		age Int
		friends [String]
	} representation tuple`

	// Create a Person instance
	person := &Person{
		Name:    "Alice",
		Age:     34,
		Friends: []string{"Bob", "Charlie"},
	}

	// Load the schema
	ts, err := ipld.LoadSchemaBytes([]byte(schema))
	if err != nil {
		panic(err)
	}

	// Get the Person type from the schema
	personType := ts.TypeByName("Person")

	// Bind our Go struct to the IPLD schema type
	// This creates an IPLD Node representing our data
	node := bindnode.Wrap(person, personType)

	// Get the representation according to the schema
	// This applies the representation strategy (tuple in this case)
	nodeRepr := node.Representation()

	// Encode to DAG-JSON and print
	fmt.Println("Encoded data (representation):")
	err = dagjson.Encode(nodeRepr, os.Stdout)
	if err != nil {
		panic(err)
	}
	fmt.Println()

	// For demonstration, create a new empty Person to decode into
	newPerson := &Person{}

	// Create a new bindnode that we'll fill with decoded data
	emptyNode := bindnode.Wrap(newPerson, personType)

	// Normally you would read from a file or network here
	// For simplicity, we'll just re-encode our existing node to get a []byte
	var buf []byte
	buf, err = ipld.Encode(nodeRepr, dagjson.Encode)
	if err != nil {
		panic(err)
	}

	// Decode from the byte array
	decoder, err := dagjson.DecodeOptions{}.DecodeBytes(buf)
	if err != nil {
		panic(err)
	}

	// Fill our empty node with the decoded data
	err = emptyNode.AssignNode(decoder)
	if err != nil {
		panic(err)
	}

	// The empty Person struct has now been populated
	fmt.Println("Decoded person:")
	fmt.Printf("Name: %s\n", newPerson.Name)
	fmt.Printf("Age: %d\n", newPerson.Age)
	fmt.Printf("Friends: %v\n", newPerson.Friends)
}
