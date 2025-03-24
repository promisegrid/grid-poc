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

// Schema definitions demonstrating evolutionary versioning with migration capabilities
const (
	schemaV1 = `
        type Person struct {
            name String
            age Int
        } representation tuple  # Tuple representation for compact storage
    `

	schemaV2 = `
        type Person struct {
            name String
            age Int
            email optional String  # Optional field added in v2
        } representation map       # Map representation for forward compatibility
    `
)

// Go type definitions mirroring schema evolution with IPLD struct tags
type PersonV1 struct { // Original schema (version 1) - tuple format
	Name string `ipld:"name"`
	Age  int64  `ipld:"age"`
}

type PersonV2 struct { // Evolved schema (version 2) - map format
	Name  string  `ipld:"name"`
	Age   int64   `ipld:"age"`
	Email *string `ipld:"email,omitempty"` // Optional field using pointer with omitempty
}

func main() {
	// Initialize schema type systems for both versions
	tsV1 := loadSchema(schemaV1, "Person")
	tsV2 := loadSchema(schemaV2, "Person")

	// Create sample data using original schema version
	personV1 := &PersonV1{
		Name: "Alice",
		Age:  30,
	}

	// Serialize data using v1 schema's tuple representation ([name, age])
	encoded := encodeData(personV1, tsV1)

	// Demonstrate IPLD's three-phase migration pattern:
	// 1. Attempt parse with latest schema (v2 map format)
	// 2. If incompatible, parse with legacy schema (v1 tuple)
	// 3. Transform data to latest format with automatic field mapping
	migratedPerson := migrateData(encoded, tsV1, tsV2)

	fmt.Printf("Migrated person: %+v\n", migratedPerson)
}

// loadSchema initializes a type system from schema text with validation
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

// encodeData serializes Go values using schema-specific representation
func encodeData(data interface{}, ts *schema.TypeSystem) []byte {
	schemaType := ts.TypeByName("Person")
	if schemaType == nil {
		log.Fatal("schema type Person not found")
	}

	node := bindnode.Wrap(data, schemaType) // Bind Go type to schema node

	var buf bytes.Buffer
	// Encode using the schema's representation (tuple for v1, map for v2)
	if err := dagjson.Encode(node.Representation(), &buf); err != nil {
		log.Fatal("encoding failed: ", err)
	}
	return buf.Bytes()
}

// migrateData implements IPLD's recommended migration strategy:
// 1. Attempt to parse with current schema (forward compatibility)
// 2. Fallback to legacy schema parsing (backward compatibility)
// 3. Transform data structure to current schema format
func migrateData(encoded []byte, oldTS, newTS *schema.TypeSystem) *PersonV2 {
	// Phase 1: Try parsing with current schema version (v2 map format)
	if personV2, ok := tryParse[PersonV2](encoded, newTS); ok {
		return personV2
	}

	// Phase 2: Fallback to legacy schema parsing (v1 tuple format)
	personV1, ok := tryParse[PersonV1](encoded, oldTS)
	if !ok {
		log.Printf("data: %s", string(encoded))
		log.Fatal("Data incompatible with all known schema versions")
	}

	// Phase 3: Transform v1 data to v2 format with auto field mapping
	return &PersonV2{
		Name:  personV1.Name,
		Age:   personV1.Age,
		Email: nil, // Explicit default for new optional field
	}
}

// tryParse demonstrates IPLD's schema-aware decoding with automatic format detection
func tryParse[T any](encoded []byte, ts *schema.TypeSystem) (*T, bool) {
	nodeType := ts.TypeByName("Person")
	if nodeType == nil {
		return nil, false
	}

	// Create prototype with schema type information for guided parsing
	proto := bindnode.Prototype((*T)(nil), nodeType)
	builder := proto.NewBuilder() // Schema-guided parser construction

	// Attempt decoding using schema expectations for representation format
	if err := dagjson.Decode(builder, bytes.NewReader(encoded)); err != nil {
		log.Printf("schema migration debug: %v", err)
		return nil, false
	}

	// Unwrap schema node to concrete Go type with type safety
	result := bindnode.Unwrap(builder.Build()).(*T)
	return result, true
}
