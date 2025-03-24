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

// Schema definitions for different versions
const (
	schemaV1 = `
        type Person struct {
            name String
            age Int
        } representation tuple
    `

	schemaV2 = `
        type Person struct {
            name String
            age Int
            email String (optional)
        } representation tuple
    `
)

// Go type definitions
type PersonV1 struct {
	Name string
	Age  int64
}

type PersonV2 struct {
	Name  string
	Age   int64
	Email *string // Optional field
}

func main() {
	// Load schema versions
	tsV1 := loadSchema(schemaV1, "Person")
	tsV2 := loadSchema(schemaV2, "Person")

	// Create sample V1 data
	personV1 := &PersonV1{
		Name: "Alice",
		Age:  30,
	}

	// Encode V1 data
	encoded := encodeData(personV1, tsV1)

	// Migration process
	migratedPerson := migrateData(encoded, tsV1, tsV2)

	fmt.Printf("Migrated person: %+v\n", migratedPerson)
}

func loadSchema(schemaText, typeName string) *schema.TypeSystem {
	ts, err := ipld.LoadSchemaBytes([]byte(schemaText))
	if err != nil {
		log.Fatalf("failed to load schema: %v", err)
	}
	return ts
}

func encodeData(data interface{}, ts *schema.TypeSystem) []byte {
	schemaType := ts.TypeByName("Person")
	node := bindnode.Wrap(data, schemaType)

	var buf bytes.Buffer
	if err := dagjson.Encode(node.Representation(), &buf); err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

func migrateData(encoded []byte, oldTS, newTS *schema.TypeSystem) *PersonV2 {
	// Try parsing with new schema first
	if personV2, ok := tryParse(encoded, newTS); ok {
		return personV2
	}

	// Fallback to old schema and migrate
	personV1, ok := tryParse(encoded, oldTS)
	if !ok {
		log.Fatal("data doesn't match any known schema")
	}

	// Migration logic
	return &PersonV2{
		Name:  personV1.Name,
		Age:   personV1.Age,
		Email: nil, // New field with default value
	}
}

func tryParse[T any](encoded []byte, ts *schema.TypeSystem) (*T, bool) {
	nodeType := ts.TypeByName("Person")
	builder := bindnode.Prototype((*T)(nil), nodeType).NewBuilder()

	if err := dagjson.Decode(builder, bytes.NewReader(encoded)); err != nil {
		return nil, false
	}

	return bindnode.Unwrap(builder.Build()).(*T), true
}
