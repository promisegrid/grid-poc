package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

// Schema definitions demonstrating evolutionary schema versioning with migration capabilities
// v1: Original tuple format - compact but rigid structure
// v2: Evolved map format - flexible with optional fields for forward compatibility
const (
	schemaV1 = `
        type Person struct {
            name String
            age Int
        } representation tuple  # Tuple representation (list) for compact storage
    `

	schemaV2 = `
        type Person struct {
            name String
            age Int
            email optional String  # Optional field added in v2 for extended functionality
        } representation map       # Map representation for structural flexibility
    `
)

// Go type definitions mirroring schema evolution with IPLD struct tags
type PersonV1 struct { // Original schema (version 1) - tuple format
	Name string `ipld:"name"`
	Age  int64  `ipld:"age"`
}

type PersonV2 struct { // Evolved schema (version 2) - map format with optional field
	Name  string  `ipld:"name"`
	Age   int64   `ipld:"age"`
	Email *string `ipld:"email,omitempty"` // Optional field using pointer semantics
}

func main() {
	// Initialize schema type systems for both versions with validation
	tsV1 := loadSchema(schemaV1, "Person")
	tsV2 := loadSchema(schemaV2, "Person")

	// Create sample data using original schema version (tuple format)
	personV1 := &PersonV1{
		Name: "Alice",
		Age:  30,
	}

	// Serialize data using v1 schema's tuple representation ([name, age])
	encoded := encodeData(personV1, tsV1, true)

	// 1. Attempt parse with latest schema (v2 map format)
	// 2. If incompatible, parse with legacy schema (v1 tuple)
	// 3. Transform data to latest format with automatic field mapping
	migratedPerson := migrateData(encoded, tsV1, tsV2)

	fmt.Printf("Migrated person: %+v\n", migratedPerson)
}

// loadSchema initializes and validates a type system from schema text
func loadSchema(schemaText, typeName string) *schema.TypeSystem {
	ts, err := ipld.LoadSchemaBytes([]byte(schemaText))
	if err != nil {
		log.Fatalf("schema loading failed: %v", err)
	}

	// Validate schema completeness before use
	if ts.TypeByName(typeName) == nil {
		log.Fatalf("schema missing required type %q", typeName)
	}
	return ts
}

// encodeData serializes Go values using schema-specific representation format
func encodeData(data interface{}, ts *schema.TypeSystem, useRepresentation bool) []byte {
	schemaType := ts.TypeByName("Person")
	if schemaType == nil {
		log.Fatal("schema type Person not found")
	}

	// Create schema-bound node with proper representation handling
	node := bindnode.Wrap(data, schemaType)

	var buf bytes.Buffer
	encoderNode := node
	if useRepresentation {
		// Use schema's declared representation (tuple for v1, map for v2)
		encoderNode = node.Representation()
	}

	if err := dagjson.Encode(encoderNode, &buf); err != nil {
		log.Fatal("encoding failed: ", err)
	}
	return buf.Bytes()
}

// migrateData implements IPLD's recommended migration strategy:
// 1. Attempt to parse with current schema (forward compatibility)
// 2. Fallback to legacy schema parsing (backward compatibility)
// 3. Transform data structure to current schema format
func migrateData(encoded []byte, oldTS, newTS *schema.TypeSystem) *PersonV2 {
	// Phase 1: Attempt parsing with current schema version (v2 map format)
	if personV2, ok := tryParse[PersonV2](encoded, newTS, true); ok {
		return personV2
	}

	// Phase 2: Fallback to legacy schema parsing (v1 tuple format)
	personV1, ok := tryParse[PersonV1](encoded, oldTS, false)
	if !ok {
		log.Printf("data: %s", string(encoded))
		log.Fatal("Data incompatible with all known schema versions")
	}

	// Phase 3: Transform v1 data to v2 format with auto field mapping
	return &PersonV2{
		Name:  personV1.Name,
		Age:   personV1.Age,
		Email: nil, // Default initialization for new optional field
	}
}

// tryParse demonstrates schema-guided decoding with automatic format detection
func tryParse[T any](encoded []byte, ts *schema.TypeSystem, useRepresentation bool) (*T, bool) {
	nodeType := ts.TypeByName("Person")
	if nodeType == nil {
		return nil, false
	}

	// Configure schema-aware decoding with representation handling
	proto := bindnode.Prototype((*T)(nil), nodeType)
	builder := proto.NewBuilder()

	decoderNode := builder
	if useRepresentation {
		// Handle schema-specific representation formats
		decoderNode = proto.Representation().NewBuilder()
	}

	// Decode using appropriate format expectations
	if err := dagjson.Decode(decoderNode, bytes.NewReader(encoded)); err != nil {
		log.Printf("schema migration debug: %v", err)
		return nil, false
	}

	// Handle representation vs. type-system node unwrapping
	finalNode := decoderNode.Build()
	if useRepresentation {
		// Convert representation node back to type system node
		finalNode = finalNode.(schema.TypedNode).Representation()
	}

	result := bindnode.Unwrap(finalNode).(*T)
	return result, true
}
