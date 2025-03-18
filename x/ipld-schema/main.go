package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
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

	// Load the schema. This loads the schema and returns a type system.
	ts, err := ipld.LoadSchemaBytes([]byte(schema))
	if err != nil {
		panic(err)
	}

	// Get the Person type from the schema
	personType := ts.TypeByName("Person")

	// Bind our Go struct to the IPLD schema type.
	// This creates an IPLD Node representing our data.
	node := bindnode.Wrap(person, personType)

	// Get the representation according to the schema.
	// This applies the representation strategy ("tuple" in this case).
	nodeRepr := node.Representation()

	// --- DAG-JSON Example ---

	fmt.Println("Encoded data (DAG-JSON representation):")
	err = dagjson.Encode(nodeRepr, os.Stdout)
	if err != nil {
		panic(err)
	}
	fmt.Println()

	// For demonstration, create a new empty Person to decode into.
	newPerson := &Person{}

	// Create a new bindnode that we'll fill with decoded data.
	emptyNode := bindnode.Wrap(newPerson, personType)

	// Normally you would read from a file or network here.
	// For simplicity, we'll just re-encode our existing node to get a []byte.
	var buf []byte
	buf, err = ipld.Encode(nodeRepr, dagjson.Encode)
	if err != nil {
		panic(err)
	}

	// Decode from the byte array.
	// Create a new builder from the empty node's representation prototype.
	builder := emptyNode.Representation().Prototype().NewBuilder()
	err = dagjson.Decode(builder, bytes.NewReader(buf))
	if err != nil {
		panic(err)
	}
	decodedNode := builder.Build()

	// Fill our empty node with the decoded data.
	newPerson, ok := bindnode.Unwrap(decodedNode).(*Person)
	if !ok {
		panic("unexpected type during DAG-JSON decoding")
	}

	// The empty Person struct has now been populated.
	fmt.Println("Decoded person (DAG-JSON):")
	fmt.Printf("Name: %s\n", newPerson.Name)
	fmt.Printf("Age: %d\n", newPerson.Age)
	fmt.Printf("Friends: %v\n", newPerson.Friends)

	// --- DAG-CBOR Example ---

	// Encode the same representation using DAG-CBOR.
	var bufCBOR []byte
	bufCBOR, err = ipld.Encode(nodeRepr, dagcbor.Encode)
	if err != nil {
		panic(err)
	}

	// For demonstration, create a new empty Person to decode DAG-CBOR content.
	newPersonCBOR := &Person{}
	emptyNodeCBOR := bindnode.Wrap(newPersonCBOR, personType)

	// Build a new node from the DAG-CBOR encoded bytes.
	builderCBOR := emptyNodeCBOR.Representation().Prototype().NewBuilder()
	err = dagcbor.Decode(builderCBOR, bytes.NewReader(bufCBOR))
	if err != nil {
		panic(err)
	}
	decodedNodeCBOR := builderCBOR.Build()

	// Unwrap the decoded node back into our Go struct.
	newPersonCBOR, ok = bindnode.Unwrap(decodedNodeCBOR).(*Person)
	if !ok {
		panic("unexpected type during DAG-CBOR decoding")
	}

	// Output the decoded Person from the DAG-CBOR encoding.
	fmt.Println("Decoded person (DAG-CBOR):")
	fmt.Printf("Name: %s\n", newPersonCBOR.Name)
	fmt.Printf("Age: %d\n", newPersonCBOR.Age)
	fmt.Printf("Friends: %v\n", newPersonCBOR.Friends)
}
