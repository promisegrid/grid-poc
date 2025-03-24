package migrations

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/node/bindnode"
)

// TestInitSchema tests that initSchema correctly loads a valid schema and that the required type is present.
func TestInitSchema(t *testing.T) {
	// Use schemaV1 from migrations.go
	ts := initSchema(schemaV1, "Person")
	if ts == nil {
		t.Fatal("initSchema returned nil for a valid schema")
	}
	if ts.TypeByName("Person") == nil {
		t.Fatal("Schema does not contain type 'Person'")
	}
}

// TestSerializePerson tests that serializePerson returns a non-empty encoded output.
func TestSerializePerson(t *testing.T) {
	ts := initSchema(schemaV1, "Person")
	original := &PersonV1{
		Name: "Bob",
		Age:  25,
	}

	encoded := serializePerson(original, ts)
	if len(encoded) == 0 {
		t.Fatal("serializePerson returned an empty result")
	}

	// Check that the encoded data contains the expected fields.
	if !strings.Contains(string(encoded), "Bob") {
		t.Errorf("Encoded data does not contain expected name, got: %s", string(encoded))
	}
	if !strings.Contains(string(encoded), "25") {
		t.Errorf("Encoded data does not contain expected age, got: %s", string(encoded))
	}
}

// TestPerformMigration tests that performMigration correctly transforms a PersonV1 into a PersonV2.
func TestPerformMigration(t *testing.T) {
	// Prepare both schema type systems.
	tsV1 := initSchema(schemaV1, "Person")
	tsV2 := initSchema(schemaV2, "Person")

	// Create a PersonV1 instance and serialize using schemaV1.
	original := &PersonV1{
		Name: "Charlie",
		Age:  40,
	}
	encoded := serializePerson(original, tsV1)

	// Perform migration.
	migrated := performMigration(encoded, tsV1, tsV2)
	if migrated == nil {
		t.Fatal("performMigration returned nil")
	}

	// Validate that the migrated data matches the original for shared fields.
	if migrated.Name != original.Name {
		t.Errorf("Expected name %q but got %q", original.Name, migrated.Name)
	}
	if migrated.Age != original.Age {
		t.Errorf("Expected age %d but got %d", original.Age, migrated.Age)
	}
	// Email is new in v2 and should be nil by default.
	if migrated.Email != nil {
		t.Errorf("Expected email to be nil but got %v", migrated.Email)
	}
}

// TestTryParseDirectV2 tests that tryParse succeeds when encoding directly with the v2 schema representation.
func TestTryParseDirectV2(t *testing.T) {
	// Prepare the v2 schema.
	ts := initSchema(schemaV2, "Person")

	// Create a PersonV2 instance with all fields.
	email := "dave@example.com"
	original := &PersonV2{
		Name:  "Dave",
		Age:   28,
		Email: &email,
	}

	// Encode using the v2 schema.
	nodeType := ts.TypeByName("Person")
	// proto := bindnode.Prototype((*PersonV2)(nil), nodeType)
	// Use map representation (v2 underlying representation)
	node := bindnode.Wrap(original, nodeType)
	var buf bytes.Buffer
	if err := dagjson.Encode(node, &buf); err != nil {
		t.Fatalf("Failed to encode PersonV2: %v", err)
	}
	encoded := buf.Bytes()

	// Try decoding directly as PersonV2.
	res, ok := tryParse[PersonV2](encoded, ts, true)
	if !ok || res == nil {
		t.Fatal("tryParse failed for a valid PersonV2 encoding")
	}
	if res.Name != original.Name || res.Age != original.Age || res.Email == nil || *res.Email != *original.Email {
		t.Errorf("Decoded value does not match original. Got %+v, expected %+v", res, original)
	}
}
