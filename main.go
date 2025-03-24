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
// v1: Original tuple format - compact but rigid structure using efficient list representation
// v2: Evolved map format - flexible key/value structure enabling forward compatibility
const (
	schemaV1 = `
        type Person struct {
            name String
            age Int
        } representation tuple  # Compact tuple storage using ordered list
    `

	schemaV2 = `
        type Person struct {
            name String
            age Int
            email optional String  # New optional field for expanded capabilities
        } representation map       # Flexible map with potential for future additions
    `
)

// Go type definitions mirroring schema evolution with IPLD struct tags
type PersonV1 struct { // Original schema version (tuple format)
	Name string `ipld:"name"` // Field order must match tuple representation
	Age  int64  `ipld:"age"`  // Explicit type mapping for schema alignment
}

type PersonV2 struct { // Evolved schema version (map format)
	Name  string  `ipld:"name"`        // Required field maintained across versions
	Age   int64   `ipld:"age"`         // Type consistency with schema Int
	Email *string `ipld:"email,omitempty"` // Optional field using pointer + omitempty
}

func main() {
	// Initialize schema systems with validation for type safety
	tsV1 := loadSchema(schemaV1, "Person")
	tsV2 := loadSchema(schemaV2, "Person")

	// Create sample data using original schema format (v1 tuple)
	personV1 := &PersonV1{
		Name: "Alice",
		Age:  30,
	}

	// Serialize data using v1's tuple representation ([name, age])
	encoded := encodeData(personV1, tsV1, true)

	// Migration process demonstrates IPLD's version resilience:
	// 1. Attempt parse with current schema (forward compatibility check)
	// 2. Fallback to original schema parsing (backward compatibility)
	// 3. Transform data structure to current format
	migratedPerson := migrateData(encoded, tsV1, tsV2)

	fmt.Printf("Migrated person: %+v\n", migratedPerson)
}

// loadSchema initializes and validates schema type system with fail-fast checks
func loadSchema(schemaText, typeName string) *schema.TypeSystem {
	ts, err := ipld.LoadSchemaBytes([]byte(schemaText))
	if err != nil {
		log.Fatalf("schema loading failed: %v", err)
	}

	// Validate schema integrity before use in critical path
	if ts.TypeByName(typeName) == nil {
		log.Fatalf("schema missing required type %q", typeName)
	}
	return ts
}

// encodeData demonstrates schema-driven serialization with representation handling
func encodeData(data interface{}, ts *schema.TypeSystem, useRepresentation bool) []byte {
	schemaType := ts.TypeByName("Person")
	if schemaType == nil {
		log.Fatal("schema type Person not found")
	}

	// Create type-bound node with schema validation
	node := bindnode.Wrap(data, schemaType)

	var buf bytes.Buffer
	encoderNode := node
	if useRepresentation {
		// Use schema-specific serialization format (tuple/map)
		encoderNode = node.Representation()
	}

	// Serialize using DAG-JSON codec with schema-enforced structure
	if err := dagjson.Encode(encoderNode, &buf); err != nil {
		log.Fatal("encoding failed: ", err)
	}
	return buf.Bytes()
}

// migrateData implements IPLD's recommended migration pattern:
// 1. Attempt modern schema parse (forward compatibility)
// 2. Fallback to legacy schema parse if needed
// 3. Transform data to modern format with field mapping
func migrateData(encoded []byte, oldTS, newTS *schema.TypeSystem) *PersonV2 {
	// Phase 1: Attempt direct parse with current schema (map format)
	if personV2, ok := tryParse[PersonV2](encoded, newTS, true); ok {
		return personV2
	}

	// Phase 2: Fallback to legacy schema parsing (tuple format)
	personV1, ok := tryParse[PersonV1](encoded, oldTS, true) // Use representation for tuple
	if !ok {
		log.Printf("Raw data: %s", string(encoded))
		log.Fatal("Data incompatible with all schema versions")
	}

	// Phase 3: Transform legacy data to current format
	return &PersonV2{
		Name:  personV1.Name,
		Age:   personV1.Age,
		Email: nil, // New fields initialized with default values
	}
}

// tryParse demonstrates schema-aware decoding with automatic format detection
func tryParse[T any](encoded []byte, ts *schema.TypeSystem, useRepresentation bool) (*T, bool) {
	nodeType := ts.TypeByName("Person")
	if nodeType == nil {
		return nil, false
	}

	// Create prototype for schema-guided decoding
	proto := bindnode.Prototype((*T)(nil), nodeType)
	builder := proto.NewBuilder()

	decoderNode := builder
	if useRepresentation {
		// Use representation-aware decoding for format-specific parsing
		decoderNode = proto.Representation().NewBuilder()
	}

	// Decode data using schema expectations
	if err := dagjson.Decode(decoderNode, bytes.NewReader(encoded)); err != nil {
		log.Printf("schema parse attempt failed: %v", err)
		return nil, false
	}

	finalNode := decoderNode.Build()
	if useRepresentation {
		// Convert representation node to type node using schema rules
		typedBuilder := proto.NewBuilder()
		if err := typedBuilder.AssignNode(finalNode); err != nil {
			log.Printf("schema conversion failed: %v", err)
			return nil, false
		}
		finalNode = typedBuilder.Build()
	}

	// Extract Go value from IPLD node with type safety
	result := bindnode.Unwrap(finalNode).(*T)
	return result, true
}
