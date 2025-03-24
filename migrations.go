package migrations

import (
	"bytes"
	"fmt"
	"log"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

// Schema definitions for demonstration purposes.
// schemaV1 uses a tuple representation (compact order-dependent format).
// schemaV2 uses a map representation (flexible key/value format with an additional optional field).
const (
	schemaV1 = `
		type Person struct {
			name String
			age  Int
		} representation tuple
	`
	schemaV2 = `
		type Person struct {
			name  String
			age   Int
			email optional String
		} representation map
	`
)

// PersonV1 corresponds to the original tuple schema.
type PersonV1 struct {
	Name string `ipld:"name"`
	Age  int64  `ipld:"age"`
}

// PersonV2 corresponds to the evolved map schema.
type PersonV2 struct {
	Name  string  `ipld:"name"`
	Age   int64   `ipld:"age"`
	Email *string `ipld:"email,omitempty"`
}

// Main is provided as an example of how to perform a migration.
// In an actual application, this might be replaced or integrated into a larger system.
func Main() {
	// Initialize the schema type systems for both versions.
	tsV1 := initSchema(schemaV1, "Person")
	tsV2 := initSchema(schemaV2, "Person")

	// Create an instance of Person using the original v1 structure.
	original := &PersonV1{
		Name: "Alice",
		Age:  30,
	}

	// Serialize the PersonV1 data using the tuple representation.
	encodedData := serializePerson(original, tsV1)
	fmt.Println("Encoded IPLD Data:", string(encodedData))

	// Migrate the data from the original format (v1) to the new format (v2).
	migrated := performMigration(encodedData, tsV1, tsV2)
	fmt.Printf("Migrated Person: %+v\n", migrated)
}

// initSchema loads and validates an IPLD schema from a schema string.
// It ensures the required type exists.
func initSchema(schemaText, typeName string) *schema.TypeSystem {
	ts, err := ipld.LoadSchemaBytes([]byte(schemaText))
	if err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}
	if ts.TypeByName(typeName) == nil {
		log.Fatalf("Schema missing required type %q", typeName)
	}
	return ts
}

// serializePerson encodes a Person using the provided schema.
// This uses the original v1 tuple representation.
func serializePerson(data interface{}, ts *schema.TypeSystem) []byte {
	nodeType := ts.TypeByName("Person")
	if nodeType == nil {
		log.Fatalf("Schema missing type Person")
	}

	// proto := bindnode.Prototype(data, nodeType)
	node := bindnode.Wrap(data, nodeType)
	reprNode := node.Representation()

	var buf bytes.Buffer
	if err := dagjson.Encode(reprNode, &buf); err != nil {
		log.Fatalf("Failed to encode person: %v", err)
	}
	return buf.Bytes()
}

// performMigration tries to decode encoded data using the new schema (v2).
// If decoding with v2 fails, it falls back to decoding with the v1 schema and then
// converting the data to the v2 structure.
func performMigration(encoded []byte, tsOld, tsNew *schema.TypeSystem) *PersonV2 {
	// Attempt decoding directly using the new schema (map representation).
	if pNew, ok := tryParse[PersonV2](encoded, tsNew, true); ok {
		return pNew
	}

	// Fallback: decode using the old schema (tuple representation).
	pOld, ok := tryParse[PersonV1](encoded, tsOld, true)
	if !ok {
		log.Fatalf("Data incompatible with both schema versions. Raw data: %s", string(encoded))
	}

	// Transform from v1 to v2; new fields get default values.
	return &PersonV2{
		Name:  pOld.Name,
		Age:   pOld.Age,
		Email: nil,
	}
}

// tryParse is a generic helper that attempts to decode IPLD-encoded data into the desired Go type.
// The useRepr flag controls whether decoding should use the schema's representation (e.g. tuple or map).
func tryParse[T any](encoded []byte, ts *schema.TypeSystem, useRepr bool) (*T, bool) {
	nodeType := ts.TypeByName("Person")
	if nodeType == nil {
		log.Println("Schema missing type Person")
		return nil, false
	}
	proto := bindnode.Prototype((*T)(nil), nodeType)
	var builder ipld.NodeBuilder
	if useRepr {
		builder = proto.Representation().NewBuilder()
	} else {
		builder = proto.NewBuilder()
	}

	if err := dagjson.Decode(builder, bytes.NewReader(encoded)); err != nil {
		log.Printf("Failed to decode using schema: %v", err)
		return nil, false
	}
	node := builder.Build()

	if useRepr {
		// Convert the representation node to its native form.
		nativeBuilder := proto.NewBuilder()
		if err := nativeBuilder.AssignNode(node); err != nil {
			log.Printf("Failed to convert representation: %v", err)
			return nil, false
		}
		node = nativeBuilder.Build()
	}

	// Unwrap the bindnode to extract the underlying Go value.
	result, ok := bindnode.Unwrap(node).(*T)
	if !ok {
		log.Println("Type assertion failed during unmarshalling")
		return nil, false
	}
	return result, true
}
