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

// Schema definitions demonstrating evolutionary versioning
const (
	schemaV1 = `
        type Person struct {
            name String
            age Int
        } representation tuple  # Compact positional mapping for v1 data
    `

	schemaV2 = `
        type Person struct {
            name String
            age Int
            email optional String  # Added optional field with forward-compatible syntax
        } representation tuple     # Maintains structural compatibility with v1
    `
)

// Go type definitions mirroring schema evolution
type PersonV1 struct {  // Original schema (version 1)
	Name string
	Age  int64
}

type PersonV2 struct {  // Evolved schema (version 2)
	Name  string
	Age   int64
	Email *string  # Nil pointer represents missing optional field (IPLD 'optional' type)
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

	// Demonstrate IPLD's recommended migration pattern:
	// 1. Attempt parse with latest schema (v2)
	// 2. If incompatible, parse with legacy schema (v1)
	// 3. Transform data to latest format
	migratedPerson := migrateData(encoded, tsV1, tsV2)

	fmt.Printf("Migrated person: %+v\n", migratedPerson)
}

// loadSchema initializes a type system from schema text
func loadSchema(schemaText, typeName string) *schema.TypeSystem {
	ts, err := ipld.LoadSchemaBytes([]byte(schemaText))
	if err != nil {
		log.Fatalf("schema loading failed: %v", err)
	}
	return ts
}

// encodeData serializes Go values using schema-specific representation
func encodeData(data interface{}, ts *schema.TypeSystem) []byte {
	schemaType := ts.TypeByName("Person")
	node := bindnode.Wrap(data, schemaType)  # Bind Go type to schema node

	var buf bytes.Buffer
	if err := dagjson.Encode(node.Representation(), &buf); err != nil {
		log.Fatal("encoding failed: ", err)
	}
	return buf.Bytes()
}

// migrateData implements IPLD's three-phase migration strategy:
// 1. Attempt to parse with latest schema version (forward compatibility)
// 2. Fallback to legacy schema parsing if needed (backward compatibility)
// 3. Transform data structure to current schema format
func migrateData(encoded []byte, oldTS, newTS *schema.TypeSystem) *PersonV2 {
	// Phase 1: Try parsing with current schema version
	if personV2, ok := tryParse[PersonV2](encoded, newTS); ok {
		return personV2
	}

	// Phase 2: Fallback to legacy schema parsing
	personV1, ok := tryParse[PersonV1](encoded, oldTS)
	if !ok {
		log.Fatal("data incompatible with all known schema versions")
	}

	// Phase 3: Transform v1 data to v2 format
	return &PersonV2{
		Name:  personV1.Name,
		Age:   personV1.Age,
		Email: nil,  # Explicit default for new optional field
	}
}

// tryParse demonstrates IPLD's schema-aware decoding pattern
func tryParse[T any](encoded []byte, ts *schema.TypeSystem) (*T, bool) {
	nodeType := ts.TypeByName("Person")
	proto := bindnode.Prototype((*T)(nil), nodeType)
	builder := proto.NewBuilder()  # Schema-guided parser construction

	if err := dagjson.Decode(builder, bytes.NewReader(encoded)); err != nil {
		return nil, false
	}

	return bindnode.Unwrap(builder.Build()).(*T), true
}
