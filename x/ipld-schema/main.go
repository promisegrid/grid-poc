package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/node/bindnode"

	// import fxamacker/cbor/v2 for CBOR encoding/decoding and diagnostic output

	"github.com/stevegt/cbordiag"
)

// Define Go structures that match our schema
type Person struct {
	Hash    []byte
	Name    string
	Age     int
	Friends []string
}

func main() {
	// --- Tuple Representation Example ---
	// Define an IPLD Schema with tuple representation.
	schema := `type Person struct {
		hash Bytes
		name String
		age Int
		friends [String]
	} representation tuple`

	// Create a Person instance.
	person := &Person{
		Hash:    []byte{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0},
		Name:    "Alice",
		Age:     34,
		Friends: []string{"Bob", "Charlie"},
	}

	// Load the tuple schema. This loads the schema and returns a type system.
	ts, err := ipld.LoadSchemaBytes([]byte(schema))
	if err != nil {
		panic(err)
	}

	// Get the Person type from the schema.
	personType := ts.TypeByName("Person")

	// Bind our Go struct to the IPLD schema type.
	// This creates an IPLD Node representing our data.
	node := bindnode.Wrap(person, personType)

	// Get the representation according to the tuple schema.
	nodeRepr := node.Representation()

	// --- DAG-JSON Example (Tuple) ---
	fmt.Println("Encoded data (DAG-JSON representation, tuple):")
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

	// Create a new builder from the empty node's representation prototype.
	builder := emptyNode.Representation().Prototype().NewBuilder()
	// Decode into the empty node from the bytes.
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
	fmt.Println("Decoded person (DAG-JSON, tuple):")
	fmt.Printf("Hash: %x\n", newPerson.Hash)
	fmt.Printf("Name: %s\n", newPerson.Name)
	fmt.Printf("Age: %d\n", newPerson.Age)
	fmt.Printf("Friends: %v\n", newPerson.Friends)

	// --- DAG-CBOR Example (Tuple) ---
	// Encode the same representation using DAG-CBOR.
	var bufCBOR []byte
	bufCBOR, err = ipld.Encode(nodeRepr, dagcbor.Encode)
	if err != nil {
		panic(err)
	}
	fmt.Println()

	// Show the encoded bytes as a hex dump
	fmt.Println("\nEncoded data (DAG-CBOR representation, LongBytes):")
	fmt.Printf("%x\n", bufCBOR)

	// display the diagnostic output of the encoded bytes
	fmt.Println("Diagnostic of Encoded data (DAG-CBOR representation, tuple):")
	annotated := cbordiag.Annotate(bufCBOR)
	fmt.Println(strings.Join(annotated, "\n"))
	fmt.Println()

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
	fmt.Println("Decoded person (DAG-CBOR, tuple):")
	fmt.Printf("Hash: %x\n", newPersonCBOR.Hash)
	fmt.Printf("Name: %s\n", newPersonCBOR.Name)
	fmt.Printf("Age: %d\n", newPersonCBOR.Age)
	fmt.Printf("Friends: %v\n", newPersonCBOR.Friends)

	// --- Map Representation Example ---
	// Define a new IPLD Schema with map representation.
	mapSchema := `type Person struct {
		hash Bytes
		name String
		age Int
		friends [String]
	} representation map`

	// Load the map schema.
	tsMap, err := ipld.LoadSchemaBytes([]byte(mapSchema))
	if err != nil {
		panic(err)
	}

	// Get the Person type from the new schema.
	personTypeMap := tsMap.TypeByName("Person")

	// Bind our Go struct to the IPLD schema type with map representation.
	nodeMap := bindnode.Wrap(person, personTypeMap)

	// Get the representation according to the map schema.
	nodeMapRepr := nodeMap.Representation()

	// --- DAG-JSON Example (Map) ---
	fmt.Println("\nEncoded data (DAG-JSON representation, map):")
	err = dagjson.Encode(nodeMapRepr, os.Stdout)
	if err != nil {
		panic(err)
	}
	fmt.Println()

	// --- DAG-CBOR Example (Map) ---
	// Encode using DAG-CBOR.
	var bufMapCBOR []byte
	bufMapCBOR, err = ipld.Encode(nodeMapRepr, dagcbor.Encode)
	if err != nil {
		panic(err)
	}

	// For demonstration, create a new empty Person to decode DAG-CBOR content.
	newPersonMap := &Person{}
	emptyNodeMap := bindnode.Wrap(newPersonMap, personTypeMap)
	builderMap := emptyNodeMap.Representation().Prototype().NewBuilder()
	err = dagcbor.Decode(builderMap, bytes.NewReader(bufMapCBOR))
	if err != nil {
		panic(err)
	}
	decodedNodeMap := builderMap.Build()
	newPersonMap, ok = bindnode.Unwrap(decodedNodeMap).(*Person)
	if !ok {
		panic("unexpected type during DAG-CBOR decoding for map representation")
	}

	// Output the decoded Person from the DAG-CBOR encoding.
	fmt.Println("Decoded person (DAG-CBOR, map):")
	fmt.Printf("Hash: %x\n", newPersonMap.Hash)
	fmt.Printf("Name: %s\n", newPersonMap.Name)
	fmt.Printf("Age: %d\n", newPersonMap.Age)
	fmt.Printf("Friends: %v\n", newPersonMap.Friends)

	// simple example of a single-element tuple containing a long byte array
	byteSchema := `type LongBytes struct {
		data Bytes
		} representation tuple`

	type LongBytes struct {
		Data []byte
	}

	// Create a LongBytes instance.
	longBytes := &LongBytes{
		Data: []byte{
			0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
			0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
			0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
			0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
		},
	}

	// Load the byte schema.
	tsByte, err := ipld.LoadSchemaBytes([]byte(byteSchema))
	if err != nil {
		panic(err)
	}

	// Get the LongBytes type from the schema.
	longBytesType := tsByte.TypeByName("LongBytes")
	// Bind our Go struct to the IPLD schema type.
	nodeLongBytes := bindnode.Wrap(longBytes, longBytesType)
	// Get the representation according to the tuple schema.
	nodeLongBytesRepr := nodeLongBytes.Representation()
	// Encode using DAG-CBOR.
	var bufLongBytesCBOR []byte
	bufLongBytesCBOR, err = ipld.Encode(nodeLongBytesRepr, dagcbor.Encode)
	if err != nil {
		panic(err)
	}

	// Show the diagnostic output of the encoded bytes
	fmt.Println("\nDiagnostic of Encoded data (DAG-CBOR representation, LongBytes):")
	annotatedLongBytes := cbordiag.Annotate(bufLongBytesCBOR)
	fmt.Println(strings.Join(annotatedLongBytes, "\n"))
	fmt.Println()

}
